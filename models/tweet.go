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
