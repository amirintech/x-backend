package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/aimrintech/x-backend/models"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type UserStore interface { 
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) (*models.User, error)
	UpdateUser(user *models.User) (*models.User, error)
	DeleteUser(id string) error
	FollowUser(followerID, followingID string) error
	UnfollowUser(followerID, followingID string) error
	GetFollowers(userID string, limit int, offset int) ([]*models.User, error)
	GetFollowing(userID string, limit int, offset int) ([]*models.User, error)
	GetUserTweets(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserReplies(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserLikes(userID string, limit int, offset int) ([]*models.Tweet, error)
	GetUserRetweets(userID string, limit int, offset int) ([]*models.Tweet, error)
}

type userStore struct {
	driver *neo4j.DriverWithContext
	dbCtx *context.Context
}

func NewUserStore(driver *neo4j.DriverWithContext, dbCtx *context.Context) UserStore {
	return &userStore{
		driver: driver,
		dbCtx: dbCtx,
	}
}

func (s *userStore) GetUserByID(id string) (*models.User, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $id}) RETURN u`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	return extractUser(res)
}

func (s *userStore) GetUserByEmail(email string) (*models.User, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {email: $email}) RETURN u`,
		map[string]any{"email": email},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	return extractUser(res)
}

func (s *userStore) CreateUser(user *models.User) (*models.User, error) {
	userID := uuid.New().String()

	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`CREATE (u:User {
			id: $id, 
			name: $name, 
			email: $email, 
			password: $password,
			username: $username,
			profilePicture: null,
			bannerPicture: null,
			createdAt: datetime(),
			updatedAt: datetime(),
			isVerified: false,
			followersCount: 0,
			followingCount: 0,
			tweetsCount: 0,
			isLocked: false,
			birthday: null,
			website: null,
			bio: null,
			location: null
		}) RETURN u`,
		map[string]any{
			"id":       userID,
			"name":     user.Name,
			"email":    user.Email,
			"password": user.Password,
			"username": user.Username,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	return extractUser(res)
}



func (s *userStore) UpdateUser(user *models.User) (*models.User, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $id}) 
		SET u.name = $name, 
			u.email = $email, 
			u.password = $password, 
			u.username = $username, 
			u.profilePicture = $profilePicture, 
			u.bannerPicture = $bannerPicture, 
			u.isVerified = $isVerified, 
			u.followersCount = $followersCount, 
			u.followingCount = $followingCount, 
			u.tweetsCount = $tweetsCount, 
			u.isLocked = $isLocked, 
			u.birthday = $birthday, 
			u.website = $website, 
			u.bio = $bio, 
			u.location = $location 
			RETURN u`,
		map[string]any{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"password":        user.Password,
			"username":        user.Username,
			"profilePicture":  user.ProfilePicture,
			"bannerPicture":   user.BannerPicture,
			"isVerified":      user.IsVerified,
			"followersCount":  user.FollowersCount,
			"followingCount":  user.FollowingCount,
			"tweetsCount":     user.TweetsCount,
			"isLocked":        user.IsLocked,
			"birthday":        user.Birthday,
			"website":         user.Website,
			"bio":             user.Bio,
			"location":        user.Location,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	return extractUser(res)
}

func (s *userStore) DeleteUser(id string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $id}) DELETE u`,
		map[string]any{"id": id},
		neo4j.EagerResultTransformer,
	)		
	if err != nil {
		return err
	}

	return nil
}

func (s *userStore) FollowUser(followerID, followingID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (f:User {id: $followerID}), (t:User {id: $followingID}) CREATE (f)-[:FOLLOWS]->(t)`,
		map[string]any{"followerID": followerID, "followingID": followingID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *userStore) UnfollowUser(followerID, followingID string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (f:User {id: $followerID})-[:FOLLOWS]->(t:User {id: $followingID}) DELETE f-[:FOLLOWS]->t`,
		map[string]any{"followerID": followerID, "followingID": followingID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *userStore) GetFollowers(userID string, limit int, offset int) ([]*models.User, error) {
	return nil, nil
}

func (s *userStore) GetFollowing(userID string, limit int, offset int) ([]*models.User, error) {
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


func extractUser(res *neo4j.EagerResult) (*models.User, error) {
	if len(res.Records) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	user, ok := res.Records[0].Get("u")
	if !ok {
		return nil, fmt.Errorf("failed to extract user node")	
	}		

	props := user.(neo4j.Node).Props

	return &models.User{
		ID: props["id"].(string),
		Name: props["name"].(string),
		Email: props["email"].(string),
		Password: props["password"].(string),
		Username: props["username"].(string),
		ProfilePicture: toStringPtr(props["profilePicture"]),
		BannerPicture: toStringPtr(props["bannerPicture"]),
		IsVerified: props["isVerified"].(bool),
		FollowersCount: int(props["followersCount"].(int64)),
		FollowingCount: int(props["followingCount"].(int64)),
		TweetsCount: int(props["tweetsCount"].(int64)),
		IsLocked: props["isLocked"].(bool),
		Bio: toStringPtr(props["bio"]),
		Location: toStringPtr(props["location"]),
		Website: toStringPtr(props["website"]),
		Birthday: toTimePtr(props["birthday"]),
		CreatedAt: props["createdAt"].(time.Time),
		UpdatedAt: props["updatedAt"].(time.Time),
	}, nil
}