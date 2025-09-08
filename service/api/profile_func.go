package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/dilcetto/wasa/service/database"
	"github.com/julienschmidt/httprouter"
)

var ErrUnauthorized = errors.New("unauthorized")

func (rt *_router) getAuthenticatedUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrUnauthorized
	}
	tokenString := authHeader[7:]

	userID, err := ParseToken(tokenString)
	if err != nil {
		return "", ErrUnauthorized
	}
	if _, err := rt.db.GetUserById(userID); err != nil {
		if errors.Is(err, database.ErrUserDoesNotExist) {
			return "", ErrUnauthorized
		}
	}
	return userID, nil
}

func (rt *_router) search_by(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	req := requests.SearchRequest{
		User:         r.URL.Query().Get("user"),
		Conversation: r.URL.Query().Get("conversation"),
	}

	if req.User == "" && req.Conversation == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}
	if !req.IsValid() {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	var (
		users         []schema.User
		conversations []schema.Conversation
		err           error
	)

	if req.User != "" {
		users, err = rt.db.SearchUserByUsername(req.User)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to search users by username")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if req.Conversation != "" {
		conversations, err = rt.db.SearchConversationByName(req.Conversation)
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to search conversations by name")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := struct {
		Users         []schema.User         `json:"users"`
		Conversations []schema.Conversation `json:"conversations"`
	}{
		Users:         users,
		Conversations: conversations,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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

	var req requests.UsernameUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !req.IsValid() {
		http.Error(w, "Username must be between 3 and 16 characters", http.StatusBadRequest)
		return
	}

	dbErr := rt.db.UpdateUsername(userID, req.Username)
	if errors.Is(dbErr, database.ErrUserDoesNotExist) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	} else if errors.Is(dbErr, database.ErrUsernameTaken) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Username already exists"})
		return
	} else if dbErr != nil {
		ctx.Logger.WithError(dbErr).Error("failed to update username")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update username"})
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

	var req requests.ProfilePhotoUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !req.IsValid() {
		http.Error(w, "Invalid photo", http.StatusBadRequest)
		return
	}

	dbErr := rt.db.UpdateUserPhoto(userID, req.Photo)
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
