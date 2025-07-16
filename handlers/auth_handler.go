package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode"

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
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validator.New().Struct(body); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := (*h.userStore).GetUserByEmail(body.Email)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		writeError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SetAuthCookie(w, token)
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Login successful"})
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
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// check if user already exists
	user, err := (*h.userStore).GetUserByEmail(body.Email)
	if err == nil && user != nil {
		writeError(w, r, http.StatusBadRequest, "User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err = (*h.userStore).CreateUser(&models.User{
		Username: body.Username,
		Name:     body.Name,
		Email:    body.Email,
		Password: string(hashedPassword),
	}, constants.AUTH_PROVIDER_CREDS)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Set token as HTTP-only cookie
	w.Header().Set("Set-Cookie", "token="+token+"; HttpOnly; SameSite=Strict; Max-Age="+strconv.Itoa(int(constants.AUTH_TOKEN_EXPIRY.Seconds())))

	// Return success response without token in body
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Registration successful"})
}

func (h *AuthHandlers) OAuthLoginHandler(authConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := authConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (h *AuthHandlers) OAuthCallbackHandler(authConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		t, err := authConfig.Exchange(context.Background(), code)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Failed to exchange code for access token")
			return
		}

		client := authConfig.Client(context.Background(), t)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Failed to exchange code for access token")
			return
		}
		defer resp.Body.Close()

		var v map[string]any
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Failed to decode response body")
			return
		}

		email, ok := v["email"].(string)
		if !ok {
			writeError(w, r, http.StatusInternalServerError, "Email not found in OAuth response")
			return
		}

		// Check if user already exists
		user, err := (*h.userStore).GetUserByEmail(email)
		if err == nil && user != nil {
			// User exists, log them in
			token, err := utils.GenerateJWT(user.ID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "Failed to generate token")
				return
			}
			utils.SetAuthCookie(w, token)
			http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)
			return
		}

		// If error is not 'no user found', return error
		if err != nil && err.Error() != "no user found" {
			writeError(w, r, http.StatusInternalServerError, "Failed to check user existence")
			return
		}

		// User does not exist, create new user
		username, err := createUsername(v["name"].(string))
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Failed to create username")
			return
		}

		user, err = (*h.userStore).CreateUser(&models.User{
			Name:     v["name"].(string),
			Email:    email,
			Username: username,
		}, constants.AUTH_PROVIDER_GOOGLE)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Failed to create user")
			return
		}

		token, err := utils.GenerateJWT(user.ID)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		valid, err := utils.VerifyJWT(token)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Failed to verify token")
			return
		}
		if !valid {
			writeError(w, r, http.StatusUnauthorized, "Invalid token")
			return
		}

		utils.SetAuthCookie(w, token)
		http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)
	}
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	utils.ClearAuthCookie(w)
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Logout successful"})
}

func createUsername(name string) (string, error) {
	username, err := nanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 10)
	if err != nil {
		return "", err
	}
	username = name + "_" + username
	username = strings.ReplaceAll(username, " ", "_")
	username = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			return r
		}
		return -1
	}, username)
	username = strings.ToLower(username)
	return username, nil
}
