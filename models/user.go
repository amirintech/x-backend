package models

import "time"

type User struct {
	ID        string `json:"id" neo4j:"id"`
	Username  string `json:"username" neo4j:"username"`
	Email     string `json:"email" neo4j:"email"`
	Password  string `json:"password" neo4j:"password"`
	CreatedAt time.Time `json:"createdAt" neo4j:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" neo4j:"updatedAt"`
	Bio       *string `json:"bio" neo4j:"bio"`
	Location  *string `json:"location" neo4j:"location"`
	Birthday  *time.Time `json:"birthday" neo4j:"birthday"`
	Website   *string `json:"website" neo4j:"website"`
	ProfilePicture *string `json:"profilePicture" neo4j:"profilePicture"`
	BannerPicture *string `json:"bannerPicture" neo4j:"bannerPicture"`
	IsVerified bool `json:"isVerified" neo4j:"isVerified"`
	FollowersCount int `json:"followersCount" neo4j:"followersCount"`
	FollowingCount int `json:"followingCount" neo4j:"followingCount"`
	TweetsCount int `json:"tweetsCount" neo4j:"tweetsCount"`
	IsLocked bool `json:"isLocked" neo4j:"isLocked"`
}