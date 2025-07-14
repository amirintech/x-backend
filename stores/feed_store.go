package stores

import (
	"context"

	"github.com/aimrintech/x-backend/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type FeedStore interface {
	GetFeed(userID string) ([]*models.Tweet, error)
}

type feedStore struct {
	db    *neo4j.DriverWithContext
	dbCtx *context.Context
}

func NewFeedStore(db *neo4j.DriverWithContext, dbCtx *context.Context) *feedStore {
	return &feedStore{db: db, dbCtx: dbCtx}
}

func (s *feedStore) GetFeed(userID string) ([]*models.Tweet, error) {
	return nil, nil
}
