package stores

import (
	"fmt"

	"github.com/aimrintech/x-backend/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type TweetStore interface {
	GetTweetByID(id string) (*models.Tweet, error)
	CreateTweet(tweet *models.Tweet) error
	UpdateTweet(tweet *models.Tweet) error
	DeleteTweet(id string) error
}

type tweetStore struct {
	driver *neo4j.DriverWithContext
}

func NewTweetStore(driver *neo4j.DriverWithContext) TweetStore {
	return &tweetStore{
		driver: driver,
	}
}

func (s *tweetStore) GetTweetByID(id string) (*models.Tweet, error) {
	fmt.Println("Getting tweet by ID: ", id)
	return nil, nil
}

func (s *tweetStore) CreateTweet(tweet *models.Tweet) error {
	fmt.Println("Creating tweet: ", tweet)
	return nil
}

func (s *tweetStore) UpdateTweet(tweet *models.Tweet) error {
	fmt.Println("Updating tweet: ", tweet)
	return nil
}

func (s *tweetStore) DeleteTweet(id string) error {
	fmt.Println("Deleting tweet: ", id)
	return nil
}