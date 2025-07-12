package stores

import (
	"context"
	"errors"

	"github.com/aimrintech/x-backend/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NotificationsStore interface {
	GetNotifications(userID string, limit int, offset int) ([]*models.Notification, error)
	CreateNotification(notification *models.Notification) error
	FlagAsRead(notificationID string, userID string) error
}

type notificationsStore struct {
	driver *neo4j.DriverWithContext
	dbCtx  *context.Context
}

func NewNotificationsStore(driver *neo4j.DriverWithContext, dbCtx *context.Context) NotificationsStore {
	return &notificationsStore{
		driver: driver,
		dbCtx:  dbCtx,
	}
}

func (s *notificationsStore) GetNotifications(userID string, limit int, offset int) ([]*models.Notification, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[:NOTIFICATIONS]->(n:Notification)
		LIMIT $limit
		OFFSET $offset
		ORDER BY n.createdAt DESC
		RETURN n`,
		map[string]any{"userID": userID, "limit": limit, "offset": offset},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	notifications := make([]*models.Notification, len(res.Records))
	for _, record := range res.Records {
		notification, ok := record.Get("n")
		if !ok {
			return nil, errors.New("notification not found")
		}
		notifications = append(notifications, notification.(*models.Notification))
	}

	return notifications, nil
}

func (s *notificationsStore) CreateNotification(notification *models.Notification) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (author:User {id: $authorUserID})
		MATCH (target:User {id: $targetUserID})
		OPTIONAL MATCH (tweet:Tweet {id: $targetTweetID})
		CREATE (n:Notification {
			id: $notificationID,
			targetUserID: $targetUserID,
			authorUserID: $authorUserID,
			type: $type,
			isRead: false,
			createdAt: datetime()
		})
		MERGE (author)-[:CREATED]->(n)
		MERGE (n)-[:TARGETED]->(target)
		FOREACH (_ IN CASE WHEN tweet IS NULL THEN [] ELSE [1] END |
			MERGE (n)-[:ON_TWEET]->(tweet)
			SET n.targetTweetID = $targetTweetID
		)
		`,
		map[string]any{
			"notificationID": notification.ID,
			"type":           notification.Type,
			"targetUserID":   notification.TargetUserID,
			"targetTweetID":  notification.TargetTweetID,
			"authorUserID":   notification.AuthorUserID,
		},
		neo4j.EagerResultTransformer,
	)

	return err
}

func (s *notificationsStore) FlagAsRead(notificationID string, userID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (n:Notification {id: $notificationID})
		WHERE n-[:TARGETED]->(u:User {id: $userID})
		SET n.isRead = true`,
		map[string]any{"notificationID": notificationID, "userID": userID},
		neo4j.EagerResultTransformer,
	)

	return err
}
