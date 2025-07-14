package constants

import "time"

type contextKey string

const USER_ID_KEY contextKey = "userID"

type AuthProvider string

const (
	AUTH_PROVIDER_GOOGLE AuthProvider = "google"
	AUTH_PROVIDER_CREDS  AuthProvider = "creds"
)

const AUTH_TOKEN_EXPIRY = 72 * time.Hour
