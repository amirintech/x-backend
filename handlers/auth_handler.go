package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aimrintech/x-backend/constants"
	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/stores"
	"github.com/aimrintech/x-backend/utils"
	"github.com/go-playground/validator/v10"
	nanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
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

	if err := validator.New().Struct(body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := (*h.userStore).GetUserByEmail(body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := utils.GenerateJWT(user.ID)
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
	}, constants.AuthProviderCreds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandlers) OAuthLoginHandler(authConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := authConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (h *AuthHandlers) OAuthCallbackHandler(authConfig *oauth2.Config) http.HandlerFunc {
	fmt.Println("OAuthCallbackHandler")
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("r", r)
		code := r.URL.Query().Get("code")

		t, err := authConfig.Exchange(context.Background(), code)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to exchange code for access token")
			return
		}

		client := authConfig.Client(context.Background(), t)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to exchange code for access token")
			return
		}
		defer resp.Body.Close()
		fmt.Println("resp", resp)

		var v map[string]any
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to decode response body")
			return
		}

		fmt.Println("v", v)

		username, err := nanoid.New(10)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to generate username")
			return
		}

		user, err := (*h.userStore).CreateUser(&models.User{
			Name:     v["name"].(string),
			Email:    v["email"].(string),
			Username: v["name"].(string) + "_" + username,
		}, constants.AuthProviderGoogle)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		token, err := utils.GenerateJWT(user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}
