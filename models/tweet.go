package models

import "time"

type Tweet struct {
	ID            string    `json:"id" neo4j:"id"`
	Content       *string   `json:"content" neo4j:"content"`
	CreatedAt     time.Time `json:"createdAt" neo4j:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" neo4j:"updatedAt"`
	MediaURLs     *[]string `json:"mediaURLs" neo4j:"mediaURLs"`
	LikesCount    int       `json:"likesCount" neo4j:"likesCount"`
	RepliesCount  int       `json:"repliesCount" neo4j:"repliesCount"`
	RetweetsCount int       `json:"retweetsCount" neo4j:"retweetsCount"`
	QuotesCount   int       `json:"quotesCount" neo4j:"quotesCount"`
	ViewsCount    int       `json:"viewsCount" neo4j:"viewsCount"`
	Hashtags      *[]string `json:"hashtags" neo4j:"hashtags"`
}

type TweetProps struct {
	CreatedAt     string   `json:"createdAt"`
	RepliesCount  int      `json:"repliesCount"`
	MediaURLs     []string `json:"mediaURLs"`
	ID            string   `json:"id"`
	RetweetsCount int      `json:"retweetsCount"`
	ViewsCount    int      `json:"viewsCount"`
	Content       string   `json:"content"`
	LikesCount    int      `json:"likesCount"`
	UpdatedAt     string   `json:"updatedAt"`
	Hashtags      []string `json:"hashtags"`
	Author        struct {
		ID             string  `json:"id"`
		IsVerified     *bool   `json:"isVerified"`
		Username       string  `json:"username"`
		ProfilePicture *string `json:"profilePicture"`
		Name           *string `json:"name"`
	} `json:"author"`
	IsLiked      bool `json:"isLiked"`
	IsRetweeted  bool `json:"isRetweeted"`
	IsBookmarked bool `json:"isBookmarked"`
}
