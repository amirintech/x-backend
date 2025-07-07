package stores

import (
	"fmt"

	"github.com/aimrintech/x-backend/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type UserStore interface { 
	GetUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
	FollowUser(followerID, followingID string) error
	UnfollowUser(followerID, followingID string) error
	GetFollowers(userID string) ([]*models.User, error)
	GetFollowing(userID string) ([]*models.User, error)
	GetUserTweets(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserReplies(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserLikes(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserRetweets(userID string, limit int, offset int) ([]*models.Tweet, error)
}

type userStore struct {
	db *neo4j.DriverWithContext
}

func NewUserStore(db *neo4j.DriverWithContext) UserStore {
	return &userStore{
		db: db,
	}
}

func (s *userStore) GetUserByID(id string) (*models.User, error) {
	fmt.Println("Getting user by ID: ", id)
	return nil, nil
}

func (s *userStore) CreateUser(user *models.User) error {
	fmt.Println("Creating user: ", user)
	return nil
}

func (s *userStore) UpdateUser(user *models.User) error {
	fmt.Println("Updating user: ", user)
	return nil
}

func (s *userStore) DeleteUser(id string) error {
	fmt.Println("Deleting user: ", id)
	return nil
}

func (s *userStore) FollowUser(followerID, followingID string) error {
	fmt.Println("Following user: ", followerID, " to ", followingID)
	return nil
}

func (s *userStore) UnfollowUser(followerID, followingID string) error {
	fmt.Println("Unfollowing user: ", followerID, " from ", followingID)
	return nil
}

func (s *userStore) GetFollowers(userID string) ([]*models.User, error) {
	fmt.Println("Getting followers for user: ", userID)
	return nil, nil
}

func (s *userStore) GetFollowing(userID string) ([]*models.User, error) {
	fmt.Println("Getting following for user: ", userID)
	return nil, nil
}

func (s *userStore) GetUserTweets(userID string, limit int, offset int) ([]*models.Tweet, error) {
	fmt.Println("Getting tweets for user: ", userID)
	return nil, nil
}

func (s *userStore) GetUserReplies(userID string, limit int, offset int) ([]*models.Tweet, error) {
	fmt.Println("Getting replies for user: ", userID)
	return nil, nil
}

func (s *userStore) GetUserLikes(userID string, limit int, offset int) ([]*models.Tweet, error) {
	fmt.Println("Getting likes for user: ", userID)
	return nil, nil
}

func (s *userStore) GetUserRetweets(userID string, limit int, offset int) ([]*models.Tweet, error) {
	fmt.Println("Getting retweets for user: ", userID)
	return nil, nil
}