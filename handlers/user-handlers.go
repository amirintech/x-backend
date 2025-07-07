package handlers

import (
	"net/http"

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

func (h *UserHandlers) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	user, err := (*h.userStore).GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	type CreateUserRequestBody struct {
		Username string `json:"username" validate:"required,min=2,max=32,alphanum"`
		Name string `json:"name" validate:"required,min=2,max=32"`
		Email string `json:"email" validate:"required,email,min=6,max=255"`
		Password string `json:"password" validate:"required,min=8,max=255"`
	}

	var body CreateUserRequestBody
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &models.User{
		Username: body.Username,
		Name: body.Name,
		Email: body.Email,
		Password: body.Password,
	}

	createdUser, err := (*h.userStore).CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, createdUser)
}