package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/database"
	"github.com/julienschmidt/httprouter"
)

var ErrUnauthorized = errors.New("unauthorized")

func (rt *_router) getAuthenticatedUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", ErrUnauthorized
	}
	userID := authHeader[7:]
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

func (rt *_router) search_by(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	query := r.URL.Query().Get("username")
	if query == "" {
		http.Error(w, "Missing username query parameter", http.StatusBadRequest)
		return
	}

	user, err := rt.db.SearchUserByName(query)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get user by name")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if user == nil {
		if err := json.NewEncoder(w).Encode([]string{}); err != nil {
			ctx.Logger.WithError(err).Error("Failed to encode empty user list")
			http.Error(w, "Failed to encode empty user list", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode user")
		http.Error(w, "Failed to encode user", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Unauthorized access")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		NewUsername string `json:"new_username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.NewUsername) < 3 || len(req.NewUsername) > 16 {
		http.Error(w, "Username must be between 3 and 16 characters", http.StatusBadRequest)
		return
	}

	dbErr := rt.db.UpdateUsername(userID, req.NewUsername)
	if errors.Is(dbErr, database.ErrUserDoesNotExist) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update username")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		ctx.Logger.WithError(err).Error("Unauthorized access")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		NewPhoto string `json:"new_photo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dbErr := rt.db.UpdateUserPhoto(userID, req.NewPhoto)
	if errors.Is(dbErr, database.ErrUserDoesNotExist) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update user photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
