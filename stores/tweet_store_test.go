package stores

import (
	"context"
	"log"
	"os"
	"path"
	"testing"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/services/notifications"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	err = godotenv.Load(path.Join(cwd, "../.env"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func setupTestTweetStore(t *testing.T) (TweetStore, *models.User, func()) {
	var (
		dbUri      = os.Getenv("NEO4J_URI")
		dbUser     = os.Getenv("NEO4J_USERNAME")
		dbPassword = os.Getenv("NEO4J_PASSWORD")
	)
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		t.Fatalf("Failed to create driver: %v", err)
	}
	err = wipeDatabase(driver)
	if err != nil {
		t.Fatalf("Failed to wipe database: %v", err)
	}
	ctx := context.Background()
	notificationsService := notifications.NewNotificationsService()
	store := NewTweetStore(&driver, &ctx, notificationsService)
	userStore := NewUserStore(&driver, &ctx, notificationsService)
	user := &models.User{
		Name:     "Tweet User",
		Email:    "tweetuser@example.com",
		Password: "hashedpassword",
		Username: "tweetuser",
	}
	createdUser, err := userStore.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	cleanup := func() {
		driver.Close(context.Background())
	}
	return store, createdUser, cleanup
}

func TestTweetStore_CRUD(t *testing.T) {
	store, user, cleanup := setupTestTweetStore(t)
	defer cleanup()

	content := "Hello, world!"
	hashtags := []string{"golang", "neo4j"}
	media := []string{"http://example.com/image.png"}
	tweet := &models.Tweet{
		Content:   &content,
		Hashtags:  &hashtags,
		MediaURLs: &media,
	}

	// Create
	created, err := store.CreateTweet(tweet, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, content, *created.Content)
	assert.ElementsMatch(t, hashtags, *created.Hashtags)
	assert.ElementsMatch(t, media, *created.MediaURLs)

	// Get by ID
	fetched, err := store.GetTweetByID(created.ID)
	assert.NoError(t, err)
	if fetched != nil {
		assert.Equal(t, created.ID, fetched.ID)
	}

	// Update
	newContent := "Updated tweet content"
	created.Content = &newContent
	updated, err := store.UpdateTweet(created, user.ID)
	assert.NoError(t, err)
	if updated != nil {
		assert.Equal(t, newContent, *updated.Content)
	}

	// Delete
	err = store.DeleteTweet(created.ID, user.ID)
	assert.NoError(t, err)

	// Get after delete
	deleted, err := store.GetTweetByID(created.ID)
	assert.Error(t, err)
	assert.Nil(t, deleted)
}

func TestTweetStore_LikeUnlike(t *testing.T) {
	store, user, cleanup := setupTestTweetStore(t)
	defer cleanup()

	content := "Like test"
	hashtags := []string{"like"}
	media := []string{}
	tweet := &models.Tweet{
		Content:   &content,
		Hashtags:  &hashtags,
		MediaURLs: &media,
	}
	created, err := store.CreateTweet(tweet, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	// Like
	err = store.LikeTweet(created.ID, user.ID)
	assert.NoError(t, err)

	// Unlike
	err = store.UnlikeTweet(created.ID, user.ID)
	assert.NoError(t, err)
}

func TestTweetStore_RetweetUnretweet(t *testing.T) {
	store, user, cleanup := setupTestTweetStore(t)
	defer cleanup()

	content := "Retweet test"
	hashtags := []string{"retweet"}
	media := []string{}
	tweet := &models.Tweet{
		Content:   &content,
		Hashtags:  &hashtags,
		MediaURLs: &media,
	}
	created, err := store.CreateTweet(tweet, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	// Retweet
	err = store.Retweet(created.ID, user.ID)
	assert.NoError(t, err)

	// Unretweet
	err = store.Unretweet(created.ID, user.ID)
	assert.NoError(t, err)
}

func TestTweetStore_QuoteTweet(t *testing.T) {
	store, user, cleanup := setupTestTweetStore(t)
	defer cleanup()

	content := "Original tweet"
	hashtags := []string{"original"}
	media := []string{}
	original := &models.Tweet{
		Content:   &content,
		Hashtags:  &hashtags,
		MediaURLs: &media,
	}
	created, err := store.CreateTweet(original, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	quoteContent := "This is a quote tweet"
	quoteHashtags := []string{"quote"}
	quoteMedia := []string{}
	quote := &models.Tweet{
		Content:   &quoteContent,
		Hashtags:  &quoteHashtags,
		MediaURLs: &quoteMedia,
	}
	quoted, err := store.QuoteTweet(created.ID, user.ID, quote)
	assert.NoError(t, err)
	assert.NotNil(t, quoted)
	assert.Equal(t, quoteContent, *quoted.Content)
}

func TestTweetStore_BookmarkUnbookmark(t *testing.T) {
	store, user, cleanup := setupTestTweetStore(t)
	defer cleanup()

	content := "Bookmark test"
	hashtags := []string{"bookmark"}
	media := []string{}
	tweet := &models.Tweet{
		Content:   &content,
		Hashtags:  &hashtags,
		MediaURLs: &media,
	}
	created, err := store.CreateTweet(tweet, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	// Bookmark
	err = store.BookmarkTweet(created.ID, user.ID)
	assert.NoError(t, err)

	// Unbookmark
	err = store.UnbookmarkTweet(created.ID, user.ID)
	assert.NoError(t, err)
}
