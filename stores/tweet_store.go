package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services/notifications"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type TweetStore interface {
	GetTweetByID(id string) (*models.Tweet, error)
	CreateTweet(tweet *models.Tweet, userID string) (*models.Tweet, error)
	UpdateTweet(tweet *models.Tweet, userID string) (*models.Tweet, error)
	DeleteTweet(tweetID string, userID string) error
	LikeTweet(tweetID string, userID string) error
	UnlikeTweet(tweetID string, userID string) error
	Retweet(tweetID string, userID string) error
	Unretweet(tweetID string, userID string) error
	QuoteTweet(originalTweetID string, userID string, quotedTweet *models.Tweet) (*models.Tweet, error)
	BookmarkTweet(tweetID string, userID string) error
	UnbookmarkTweet(tweetID string, userID string) error
	// ReplyToTweet(tweetID string, userID string, content string) (*models.Tweet, error)
	// GetReplies(tweetID string, limit int, offset int) ([]*models.Tweet, error)
}

type tweetStore struct {
	driver               *neo4j.DriverWithContext
	dbCtx                *context.Context
	notificationsService notifications.Notifications
}

func NewTweetStore(driver *neo4j.DriverWithContext, dbCtx *context.Context, notificationsService notifications.Notifications) TweetStore {
	return &tweetStore{
		driver:               driver,
		dbCtx:                dbCtx,
		notificationsService: notificationsService,
	}
}

func (s *tweetStore) GetTweetByID(id string) (*models.Tweet, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (t:Tweet {id: $id}) RETURN t`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	tweet, err := extractTweetFromEagerResult(res)
	if err != nil {
		return nil, err
	}

	return tweet, nil
}

func (s *tweetStore) CreateTweet(tweet *models.Tweet, userID string) (*models.Tweet, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`
		MATCH (u:User {id: $userID})
		WITH u
		CREATE (t:Tweet {
			id: $id,
			content: $content,
			createdAt: datetime(),
			updatedAt: datetime(),
			likesCount: 0,
			repliesCount: 0,
			retweetsCount: 0,
			quotesCount: 0,
			viewsCount: 0,
			hashtags: $hashtags,
			mediaURLs: $mediaURLs
		})
		MERGE (u)-[:TWEETS]->(t)
		RETURN t
		`,
		map[string]any{"id": uuid.New().String(), "content": tweet.Content, "hashtags": tweet.Hashtags, "mediaURLs": tweet.MediaURLs, "userID": userID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	tweet, err = extractTweetFromEagerResult(res)
	if err != nil {
		return nil, err
	}

	return tweet, nil
}

func (s *tweetStore) UpdateTweet(tweet *models.Tweet, userID string) (*models.Tweet, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (t:Tweet {id: $id}) SET t.content = $content, t.updatedAt = datetime() RETURN t`,
		map[string]any{"id": tweet.ID, "content": tweet.Content},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	tweet, err = extractTweetFromEagerResult(res)
	if err != nil {
		return nil, err
	}

	return tweet, nil
}

func (s *tweetStore) DeleteTweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (t:Tweet {id: $id}) DETACH DELETE t`,
		map[string]any{"id": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *tweetStore) LikeTweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`
		MATCH (u:User {id: $userID}), (t:Tweet {id: $tweetID})
		MERGE (u)-[l:LIKES]->(t)
		SET t.likesCount = t.likesCount + 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	s.notificationsService.Publish(models.NotificationTypeLike, models.NewNotification(userID, tweetID, nil, models.NotificationTypeLike))

	return nil
}

func (s *tweetStore) UnlikeTweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[l:LIKES]->(t:Tweet {id: $tweetID}) DELETE l
		SET t.likesCount = t.likesCount - 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *tweetStore) Retweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID}), (t:Tweet {id: $tweetID})
		MERGE (u)-[r:RETWEETS]->(t)
		SET t.retweetsCount = t.retweetsCount + 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	s.notificationsService.Publish(models.NotificationTypeRetweet, models.NewNotification(userID, tweetID, nil, models.NotificationTypeRetweet))

	return nil
}

func (s *tweetStore) Unretweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[r:RETWEETS]->(t:Tweet {id: $tweetID}) DELETE r
		SET t.retweetsCount = t.retweetsCount - 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *tweetStore) QuoteTweet(originalTweetID string, userID string, quotedTweet *models.Tweet) (*models.Tweet, error) {
	// create the new tweet (the quote tweet)
	createdTweet, err := s.CreateTweet(quotedTweet, userID)
	if err != nil {
		return nil, err
	}

	// create a QUOTES relationship from the new tweet to the original tweet
	_, err = neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (qt:Tweet {id: $quoteTweetID}), (ot:Tweet {id: $originalTweetID})
		MERGE (qt)-[:QUOTES]->(ot)
		SET ot.retweetsCount = ot.retweetsCount + 1
		`,
		map[string]any{"quoteTweetID": createdTweet.ID, "originalTweetID": originalTweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	s.notificationsService.Publish(models.NotificationTypeRetweet, models.NewNotification(userID, originalTweetID, nil, models.NotificationTypeRetweet))

	return createdTweet, nil
}

func (s *tweetStore) BookmarkTweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID}), (t:Tweet {id: $tweetID})
		MERGE (u)-[b:BOOKMARKS]->(t)
		SET t.bookmarksCount = t.bookmarksCount + 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *tweetStore) UnbookmarkTweet(tweetID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[b:BOOKMARKS]->(t:Tweet {id: $tweetID}) DELETE b
		SET t.bookmarksCount = t.bookmarksCount - 1
		`,
		map[string]any{"userID": userID, "tweetID": tweetID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func extractTweetFromNode(tweetNode any) *models.Tweet {
	props := tweetNode.(neo4j.Node).Props

	// convert hashtags
	var hashtags []string
	if raw, ok := props["hashtags"]; ok && raw != nil {
		for _, v := range raw.([]any) {
			hashtags = append(hashtags, v.(string))
		}
	}

	// convert mediaURLs
	var mediaURLs []string
	if raw, ok := props["mediaURLs"]; ok && raw != nil {
		for _, v := range raw.([]any) {
			mediaURLs = append(mediaURLs, v.(string))
		}
	}

	return &models.Tweet{
		ID:            props["id"].(string),
		Content:       toStringPtr(props["content"]),
		CreatedAt:     props["createdAt"].(time.Time),
		UpdatedAt:     props["updatedAt"].(time.Time),
		LikesCount:    int(props["likesCount"].(int64)),
		RepliesCount:  int(props["repliesCount"].(int64)),
		RetweetsCount: int(props["retweetsCount"].(int64)),
		QuotesCount:   int(props["quotesCount"].(int64)),
		ViewsCount:    int(props["viewsCount"].(int64)),
		Hashtags:      &hashtags,
		MediaURLs:     &mediaURLs,
	}
}

func extractTweetFromEagerResult(res *neo4j.EagerResult) (*models.Tweet, error) {
	if len(res.Records) == 0 {
		return nil, fmt.Errorf("no tweet found")

	}

	tweet, ok := res.Records[0].Get("t")
	if !ok {
		return nil, fmt.Errorf("failed to extract tweet node")
	}

	return extractTweetFromNode(tweet), nil
}
