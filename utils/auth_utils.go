package utils

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/aimrintech/x-backend/constants"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string) (string, error) {
	jwtSecret := loadJWTSecret()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(constants.AUTH_TOKEN_EXPIRY).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return loadJWTSecret(), nil
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

func VerifyJWT(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return loadJWTSecret(), nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	// w.Header().Set("Set-Cookie", "token="+token+"; HttpOnly; SameSite=Strict; Max-Age="+strconv.Itoa(int(constants.AUTH_TOKEN_EXPIRY.Seconds())))
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(constants.AUTH_TOKEN_EXPIRY.Seconds()),
	})

}

func GetAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func loadJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
