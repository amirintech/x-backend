package handlers

import (
	"net/http"

	"github.com/aimrintech/x-backend/stores"
)

type UserHandlers struct {
	userStore *stores.UserStore
}

func NewUserHandlers(userStore *stores.UserStore) *UserHandlers {
	return &UserHandlers{
		userStore: userStore,
	}
}

func (h *UserHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	
}	