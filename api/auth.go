package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aimrintech/x-backend/constants"
	"github.com/golang-jwt/jwt/v5"
)


var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("user_id not found in token")
		}
		return userID, nil
	}
	return "", errors.New("invalid token")
} 

func jwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), constants.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
} 