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

	return extractUserFromEagerResult(res)
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

	return extractUserFromEagerResult(res)
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

	return extractUserFromEagerResult(res)
}

func (s *userStore) UpdateUser(user *models.User) (*models.User, error) {
	// Build SET clauses and params dynamically
	setClauses := []string{}
	params := map[string]any{
		"id": user.ID,
	}

	if user.Name != "" {
		setClauses = append(setClauses, "u.name = $name")
		params["name"] = user.Name
	}
	if user.Email != "" {
		setClauses = append(setClauses, "u.email = $email")
		params["email"] = user.Email
	}
	if user.Password != "" {
		setClauses = append(setClauses, "u.password = $password")
		params["password"] = user.Password
	}
	if user.Username != "" {
		setClauses = append(setClauses, "u.username = $username")
		params["username"] = user.Username
	}
	if user.ProfilePicture != nil {
		setClauses = append(setClauses, "u.profilePicture = $profilePicture")
		params["profilePicture"] = user.ProfilePicture
	}
	if user.BannerPicture != nil {
		setClauses = append(setClauses, "u.bannerPicture = $bannerPicture")
		params["bannerPicture"] = user.BannerPicture
	}
	if user.IsVerified {
		setClauses = append(setClauses, "u.isVerified = $isVerified")
		params["isVerified"] = user.IsVerified
	}
	if user.FollowersCount != 0 {
		setClauses = append(setClauses, "u.followersCount = $followersCount")
		params["followersCount"] = user.FollowersCount
	}
	if user.FollowingCount != 0 {
		setClauses = append(setClauses, "u.followingCount = $followingCount")
		params["followingCount"] = user.FollowingCount
	}
	if user.TweetsCount != 0 {
		setClauses = append(setClauses, "u.tweetsCount = $tweetsCount")
		params["tweetsCount"] = user.TweetsCount
	}
	if user.IsLocked {
		setClauses = append(setClauses, "u.isLocked = $isLocked")
		params["isLocked"] = user.IsLocked
	}
	if user.Birthday != nil {
		setClauses = append(setClauses, "u.birthday = $birthday")
		params["birthday"] = user.Birthday
	}
	if user.Website != nil {
		setClauses = append(setClauses, "u.website = $website")
		params["website"] = user.Website
	}
	if user.Bio != nil {
		setClauses = append(setClauses, "u.bio = $bio")
		params["bio"] = user.Bio
	}
	if user.Location != nil {
		setClauses = append(setClauses, "u.location = $location")
		params["location"] = user.Location
	}

	if len(setClauses) == 0 {
		// Nothing to update
		return s.GetUserByID(user.ID)
	}

	query := fmt.Sprintf("MATCH (u:User {id: $id}) SET %s RETURN u", joinClauses(setClauses, ", "))

	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		query,
		params,
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	u, err := extractUserFromEagerResult(res)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *userStore) DeleteUser(id string) error {
	_, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $id}) DETACH DELETE u`,
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
		`MATCH (f:User {id: $followerID})-[r:FOLLOWS]->(t:User {id: $followingID}) DELETE r`,
		map[string]any{"followerID": followerID, "followingID": followingID},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *userStore) GetFollowers(userID string, limit int, offset int) ([]*models.User, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (f:User)-[:FOLLOWS]->(u:User {id: $userID}) LIMIT $limit OFFSET $offset RETURN f`,
		map[string]any{"userID": userID, "limit": limit, "offset": offset},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		return nil, err
	}

	users := []*models.User{}
	for _, record := range res.Records {
		userNode, ok := record.Get("f")
		if !ok {
			return nil, fmt.Errorf("failed to extract user node")
		}
		user := extractUserFromNode(userNode)
		user.Password = ""
		users = append(users, user)
	}
	return users, nil
}

func (s *userStore) GetFollowing(userID string, limit int, offset int) ([]*models.User, error) {
	res, err := neo4j.ExecuteQuery(
		*s.dbCtx,
		*s.driver,
		`MATCH (u:User {id: $userID})-[:FOLLOWS]->(f:User) LIMIT $limit OFFSET $offset RETURN f`,
		map[string]any{"userID": userID, "limit": limit, "offset": offset},
		neo4j.EagerResultTransformer,
	)	
	if err != nil {
		return nil, err
	}

	users := []*models.User{}
	for _, record := range res.Records {
		userNode, ok := record.Get("f")
		if !ok {
			return nil, fmt.Errorf("failed to extract user node")
		}
		user := extractUserFromNode(userNode)
		user.Password = ""
		users = append(users, user)
	}
	return users, nil
}

func extractUserFromNode(userNode any) (*models.User) {
	props := userNode.(neo4j.Node).Props

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
	}
}

func extractUserFromEagerResult(res *neo4j.EagerResult) (*models.User, error) {
	if len(res.Records) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	user, ok := res.Records[0].Get("u")
	if !ok {
		return nil, fmt.Errorf("failed to extract user node")	
	}		

	return extractUserFromNode(user), nil
}


func joinClauses(clauses []string, sep string) string {
	result := ""
	for i, c := range clauses {
		result += c
		if i < len(clauses)-1 {
			result += sep
		}
	}
	return result
}