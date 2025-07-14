package stores

import (
	"context"
	"log"
	"os"
	"path"
	"testing"

	"github.com/aimrintech/x-backend/constants"
	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

func init() {
	// load environment variables
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	err = godotenv.Load(path.Join(cwd, "../.env"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		panic(err)
	}
}

// deletes all nodes and relationships in the database
func wipeDatabase(driver neo4j.DriverWithContext) error {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(context.Background())
	_, err := session.Run(context.Background(), "MATCH (n) DETACH DELETE n", nil)
	return err
}

func setupTestStore(t *testing.T) (UserStore, func()) {
	var (
		dbUri      = os.Getenv("NEO4J_URI")
		dbUser     = os.Getenv("NEO4J_USERNAME")
		dbPassword = os.Getenv("NEO4J_PASSWORD")
	)
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	// wipe the database before each test
	err = wipeDatabase(driver)
	if err != nil {
		t.Fatalf("Failed to wipe database: %v", err)
	}
	ctx := context.Background()
	notificationsService := services.NewNotificationsService()
	store := NewUserStore(&driver, &ctx, notificationsService)
	cleanup := func() {
		driver.Close(context.Background())
	}
	return store, cleanup
}

func TestUserStore_CRUD(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	user := &models.User{
		Name:     "Test User",
		Email:    "testuser@example.com",
		Password: "hashedpassword",
		Username: "testuser",
	}

	// Create
	created, err := store.CreateUser(user, constants.AUTH_PROVIDER_CREDS)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, user.Email, created.Email)

	// Get by ID
	fetched, err := store.GetUserByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	// Get by Email
	fetchedByEmail, err := store.GetUserByEmail(user.Email)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, fetchedByEmail.ID)

	// Update
	created.Name = "Updated Name"
	updated, err := store.UpdateUser(created)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)

	// Delete
	err = store.DeleteUser(created.ID)
	assert.NoError(t, err)

	// Get after delete
	_, err = store.GetUserByID(created.ID)
	assert.Error(t, err)
}

func TestUserStore_FollowUnfollow(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	userA := &models.User{
		Name:     "User A",
		Email:    "usera@example.com",
		Password: "passA",
		Username: "usera",
	}
	userB := &models.User{
		Name:     "User B",
		Email:    "userb@example.com",
		Password: "passB",
		Username: "userb",
	}
	createdA, _ := store.CreateUser(userA, constants.AUTH_PROVIDER_CREDS)
	createdB, _ := store.CreateUser(userB, constants.AUTH_PROVIDER_CREDS)

	// Follow
	err := store.FollowUser(createdA.ID, createdB.ID)
	assert.NoError(t, err)

	// Get Following
	following, err := store.GetFollowing(createdA.ID, 10, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, following)

	// Get Followers
	followers, err := store.GetFollowers(createdB.ID, 10, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, followers)

	// Unfollow
	err = store.UnfollowUser(createdA.ID, createdB.ID)
	assert.NoError(t, err)

	// Clean up
	_ = store.DeleteUser(createdA.ID)
	_ = store.DeleteUser(createdB.ID)
}
