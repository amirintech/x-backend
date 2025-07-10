package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/stores"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandlers struct {
	userStore *stores.UserStore
}

func NewAuthHandlers(userStore *stores.UserStore) *AuthHandlers {
	return &AuthHandlers{userStore: userStore}
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	type LoginRequestBody struct {
		Email    string `json:"email" validate:"required,email,min=6,max=255"`
		Password string `json:"password" validate:"required,min=8,max=255"`
	}

	var body LoginRequestBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := (*h.userStore).GetUserByEmail(body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	type RegisterRequestBody struct {
		Username string `json:"username" validate:"required,min=2,max=32,alphanum"`
		Name     string `json:"name" validate:"required,min=2,max=32"`
		Email    string `json:"email" validate:"required,email,min=6,max=255"`
		Password string `json:"password" validate:"required,min=8,max=255"`
	}

	var body RegisterRequestBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := (*h.userStore).CreateUser(&models.User{
		Username: body.Username,
		Name:     body.Name,
		Email:    body.Email,
		Password: string(hashedPassword),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// func validateJWT(tokenString string) (string, error) {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, errors.New("unexpected signing method")
// 		}
// 		return jwtSecret, nil
// 	})
// 	if err != nil {
// 		return "", err
// 	}
// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		userID, ok := claims["user_id"].(string)
// 		if !ok {
// 			return "", errors.New("user_id not found in token")
// 		}
// 		return userID, nil
// 	}
// 	return "", errors.New("invalid token")
// }

// type contextKey string

// const UserIDKey contextKey = "userID"

// func jwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
// 			return
// 		}
// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 		userID, err := validateJWT(tokenString)
// 		if err != nil {
// 			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
// 			return
// 		}
// 		ctx := context.WithValue(r.Context(), UserIDKey, userID)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
