package handlers

import (
	"net/http"
	"time"

	"github.com/aimrintech/x-backend/models"
	"github.com/aimrintech/x-backend/stores"
	"github.com/go-playground/validator/v10"
)

type UserHandlers struct {
	userStore *stores.UserStore
}

func NewUserHandlers(userStore *stores.UserStore) *UserHandlers {
	return &UserHandlers{
		userStore: userStore,
	}
}

func (h *UserHandlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, err := (*h.userStore).GetUserByID(userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""
	writeJSON(w, r, http.StatusOK, user)
}

func (h *UserHandlers) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := (*h.userStore).GetUserByID(userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""
	writeJSON(w, r, http.StatusOK, user)
}

func (h *UserHandlers) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeError(w, r, http.StatusBadRequest, "Username is required")
		return
	}

	user, err := (*h.userStore).GetUserByUsername(username)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to get user")
		return
	}
	user.Password = ""
	writeJSON(w, r, http.StatusOK, user)
}

// func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
// 	type CreateUserRequestBody struct {
// 		Username string `json:"username" validate:"required,min=2,max=32,alphanum"`
// 		Name     string `json:"name" validate:"required,min=2,max=32"`
// 		Email    string `json:"email" validate:"required,email,min=6,max=255"`
// 		Password string `json:"password" validate:"required,min=8,max=255"`
// 	}

// 	var body CreateUserRequestBody
// 	if err := readJSON(r, &body); err != nil {
// 		writeError(w, http.StatusBadRequest, "Invalid request body")
// 		return
// 	}

// 	if err := validator.New().Struct(body); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	user := &models.User{
// 		Username: body.Username,
// 		Name:     body.Name,
// 		Email:    body.Email,
// 		Password: body.Password,
// 	}

// 	createdUser, err := (*h.userStore).CreateUser(user)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	writeJSON(w, http.StatusCreated, createdUser)
// }

func (h *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	type UpdateUserRequestBody struct {
		Name           string     `json:"name" validate:"omitempty,min=2,max=32"`
		ProfilePicture *string    `json:"profilePicture" validate:"omitempty,url"`
		BannerPicture  *string    `json:"bannerPicture" validate:"omitempty,url"`
		Bio            *string    `json:"bio" validate:"omitempty,max=255"`
		Location       *string    `json:"location" validate:"omitempty,max=255"`
		Website        *string    `json:"website" validate:"omitempty,url"`
		Birthday       *time.Time `json:"birthday" validate:"omitempty"`
		FollowersCount int        `json:"followersCount" validate:"omitempty"`
		FollowingCount int        `json:"followingCount" validate:"omitempty"`
		TweetsCount    int        `json:"tweetsCount" validate:"omitempty"`
	}

	var body UpdateUserRequestBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := (*h.userStore).UpdateUser(&models.User{
		ID:             userID,
		Name:           body.Name,
		ProfilePicture: body.ProfilePicture,
		BannerPicture:  body.BannerPicture,
		Bio:            body.Bio,
		Location:       body.Location,
		Website:        body.Website,
		Birthday:       body.Birthday,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Failed to update user")
		return
	}

	user.Password = ""
	writeJSON(w, r, http.StatusOK, user)
}
