package stores

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type TweetStore interface {
	GetTweets(limit int, offset int) ([]*models.Tweet, error)

	GetTweetByID(id string) (*models.Tweet, error)
	CreateTweet(tweet *models.Tweet, userID string) (*models.TweetProps, error)
	UpdateTweet(tweet *models.Tweet, userID string) (*models.TweetProps, error)
	DeleteTweet(tweetID string, userID string) error
	LikeTweet(tweetID string, userID string) error
	UnlikeTweet(tweetID string, userID string) error
	Retweet(tweetID string, userID string) error
	Unretweet(tweetID string, userID string) error
	QuoteTweet(originalTweetID string, userID string, quotedTweet *models.Tweet) (*models.TweetProps, error)
	BookmarkTweet(tweetID string, userID string) error
	UnbookmarkTweet(tweetID string, userID string) error
	GetUsersWithTweets(currUserID string, limit int, offset int) ([]models.TweetProps, error)
	// ReplyToTweet(tweetID string, userID string, content string) (*models.Tweet, error)
	// GetReplies(tweetID string, limit int, offset int) ([]*models.Tweet, error)
}

type tweetStore struct {
	driver               *neo4j.DriverWithContext
	dbCtx                *context.Context
	notificationsService services.Notifications
	feedService          services.Feed
}

func NewTweetStore(driver *neo4j.DriverWithContext, dbCtx *context.Context, notificationsService services.Notifications, feedService services.Feed) TweetStore {
	return &tweetStore{
		driver:               driver,
		dbCtx:                dbCtx,
		notificationsService: notificationsService,
		feedService:          feedService,
	}
}

func (s *tweetStore) GetTweets(limit int, offset int) ([]*models.Tweet, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (t:Tweet) LIMIT $limit OFFSET $offset RETURN t`,
		map[string]any{"limit": limit, "offset": offset},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	tweets := make([]*models.Tweet, 0, limit)
	for _, record := range res.Records {
		tweet, ok := record.Get("t")
		if !ok {
			continue
		}

		tweets = append(tweets, extractTweetFromNode(tweet))
	}

	return tweets, nil
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

func (s *tweetStore) CreateTweet(tweet *models.Tweet, userID string) (*models.TweetProps, error) {
	hashtags := extractHashtagsFromContent(*tweet.Content)
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
		RETURN t, u
		`,
		map[string]any{"id": uuid.New().String(), "content": tweet.Content, "hashtags": hashtags, "mediaURLs": tweet.MediaURLs, "userID": userID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	if len(res.Records) == 0 {
		return nil, fmt.Errorf("no tweet created")
	}

	tweetNode, okT := res.Records[0].Get("t")
	userNode, okU := res.Records[0].Get("u")
	if !okT || !okU {
		return nil, fmt.Errorf("failed to extract tweet or user node")
	}

	createdTweet := extractTweetFromNode(tweetNode)
	user := extractUserFromNode(userNode)

	// Convert to TweetProps using utility function
	tweetProps := convertTweetToProps(createdTweet, user, false, false, false)

	// Publish feed event for tweet creation (to author only for now)
	if s.feedService != nil {
		event := &models.FeedEvent{
			Type:      models.FeedEventCreated,
			Tweet:     *tweetProps,
			ActorID:   userID,
			CreatedAt: createdTweet.CreatedAt,
		}
		s.feedService.PublishToAll(event)
	}
	return tweetProps, nil
}

func (s *tweetStore) UpdateTweet(tweet *models.Tweet, userID string) (*models.TweetProps, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[:TWEETS]->(t:Tweet {id: $id}) 
		SET t.content = $content, t.updatedAt = datetime() 
		RETURN t, u`,
		map[string]any{"id": tweet.ID, "content": tweet.Content, "userID": userID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	if len(res.Records) == 0 {
		return nil, fmt.Errorf("no tweet found or user not authorized")
	}

	tweetNode, okT := res.Records[0].Get("t")
	userNode, okU := res.Records[0].Get("u")
	if !okT || !okU {
		return nil, fmt.Errorf("failed to extract tweet or user node")
	}

	updatedTweet := extractTweetFromNode(tweetNode)
	user := extractUserFromNode(userNode)

	// Convert to TweetProps using utility function
	tweetProps := convertTweetToProps(updatedTweet, user, false, false, false)

	return tweetProps, nil
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
	// Publish feed event for like (to tweet author)
	if s.feedService != nil {
		tweetProps, err := s.getTweetPropsWithUser(tweetID, userID)
		if err == nil && tweetProps != nil {
			event := &models.FeedEvent{
				Type:      models.FeedEventLiked,
				Tweet:     *tweetProps,
				ActorID:   userID,
				CreatedAt: time.Now(),
			}
			// Send to tweet author
			s.feedService.PublishToAll(event)
		}
	}

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
	// Publish feed event for retweet (to tweet author)
	if s.feedService != nil {
		tweetProps, err := s.getTweetPropsWithUser(tweetID, userID)
		if err == nil && tweetProps != nil {
			event := &models.FeedEvent{
				Type:      models.FeedEventRetweeted,
				Tweet:     *tweetProps,
				ActorID:   userID,
				CreatedAt: time.Now(),
			}
			// Send to tweet author
			s.feedService.PublishToAll(event)
		}
	}

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

func (s *tweetStore) QuoteTweet(originalTweetID string, userID string, quotedTweet *models.Tweet) (*models.TweetProps, error) {
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

func (s *tweetStore) GetUsersWithTweets(currUserID string, limit int, offset int) ([]models.TweetProps, error) {
	query := `
		MATCH (u:User)-[:TWEETS]->(t:Tweet)
		OPTIONAL MATCH (curr:User {id: $currUserID})
		OPTIONAL MATCH (curr)-[l:LIKES]->(t)
		OPTIONAL MATCH (curr)-[r:RETWEETS]->(t)
		OPTIONAL MATCH (curr)-[b:BOOKMARKS]->(t)
		WITH u, t, l, r, b
		ORDER BY t.createdAt DESC
		SKIP $offset LIMIT $limit
		RETURN u, t, l IS NOT NULL AS isLiked, r IS NOT NULL AS isRetweeted, b IS NOT NULL AS isBookmarked
	`
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		query,
		map[string]any{"currUserID": currUserID, "limit": limit, "offset": offset},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	result := make([]models.TweetProps, 0, limit)

	for _, record := range res.Records {
		userNode, okU := record.Get("u")
		tweetNode, okT := record.Get("t")
		isLiked, _ := record.Get("isLiked")
		isRetweeted, _ := record.Get("isRetweeted")
		isBookmarked, _ := record.Get("isBookmarked")
		if !okU || !okT {
			continue
		}
		user := extractUserFromNode(userNode)
		tweet := extractTweetFromNode(tweetNode)

		// Use utility function to convert to TweetProps
		tp := convertTweetToProps(tweet, user, isLiked.(bool), isRetweeted.(bool), isBookmarked.(bool))
		result = append(result, *tp)
	}

	return result, nil
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

// convertTweetToProps converts a models.Tweet and models.User to models.TweetProps
func convertTweetToProps(tweet *models.Tweet, user *models.User, isLiked, isRetweeted, isBookmarked bool) *models.TweetProps {
	tweetProps := &models.TweetProps{
		CreatedAt:     tweet.CreatedAt.Format(time.RFC3339),
		RepliesCount:  tweet.RepliesCount,
		MediaURLs:     []string{},
		ID:            tweet.ID,
		RetweetsCount: tweet.RetweetsCount,
		ViewsCount:    tweet.ViewsCount,
		Content:       "",
		LikesCount:    tweet.LikesCount,
		UpdatedAt:     tweet.UpdatedAt.Format(time.RFC3339),
		Hashtags:      []string{},
		Author: struct {
			ID             string  `json:"id"`
			IsVerified     *bool   `json:"isVerified"`
			Username       string  `json:"username"`
			ProfilePicture *string `json:"profilePicture"`
			Name           *string `json:"name"`
		}{
			ID:             user.ID,
			IsVerified:     &user.IsVerified,
			Username:       user.Username,
			ProfilePicture: user.ProfilePicture,
			Name:           &user.Name,
		},
		IsLiked:      isLiked,
		IsRetweeted:  isRetweeted,
		IsBookmarked: isBookmarked,
	}

	if tweet.Content != nil {
		tweetProps.Content = *tweet.Content
	}
	if tweet.Hashtags != nil {
		tweetProps.Hashtags = *tweet.Hashtags
	}
	if tweet.MediaURLs != nil {
		tweetProps.MediaURLs = *tweet.MediaURLs
	}

	return tweetProps
}

// getTweetPropsWithUser gets a tweet by ID and converts it to TweetProps by also fetching user information
func (s *tweetStore) getTweetPropsWithUser(tweetID string, currentUserID string) (*models.TweetProps, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User)-[:TWEETS]->(t:Tweet {id: $tweetID})
		OPTIONAL MATCH (curr:User {id: $currentUserID})
		OPTIONAL MATCH (curr)-[l:LIKES]->(t)
		OPTIONAL MATCH (curr)-[r:RETWEETS]->(t)
		OPTIONAL MATCH (curr)-[b:BOOKMARKS]->(t)
		RETURN u, t, l IS NOT NULL AS isLiked, r IS NOT NULL AS isRetweeted, b IS NOT NULL AS isBookmarked`,
		map[string]any{"tweetID": tweetID, "currentUserID": currentUserID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	if len(res.Records) == 0 {
		return nil, fmt.Errorf("tweet not found")
	}

	userNode, okU := res.Records[0].Get("u")
	tweetNode, okT := res.Records[0].Get("t")
	isLiked, _ := res.Records[0].Get("isLiked")
	isRetweeted, _ := res.Records[0].Get("isRetweeted")
	isBookmarked, _ := res.Records[0].Get("isBookmarked")

	if !okU || !okT {
		return nil, fmt.Errorf("failed to extract tweet or user node")
	}

	user := extractUserFromNode(userNode)
	tweet := extractTweetFromNode(tweetNode)

	return convertTweetToProps(tweet, user, isLiked.(bool), isRetweeted.(bool), isBookmarked.(bool)), nil
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

func extractHashtagsFromContent(content string) []string {
	re := regexp.MustCompile(`#(\w+)`)
	matches := re.FindAllString(content, -1)
	return matches
}

func extractMentionsFromContent(content string) []string {
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllString(content, -1)
	return matches
}
