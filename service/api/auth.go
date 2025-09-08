package api

import (
	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/dilcetto/wasa/service/database"

	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	w.Header().Set("Content-Type", "application/json")

	var req requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserByName(req.Username)
	if errors.Is(err, database.ErrUserDoesNotExist) {
		newID, err := generateNewID()
		if err != nil {
			http.Error(w, "Failed to generate user ID", http.StatusInternalServerError)
			return
		}
		newUser := schema.User{
			ID:       newID,
			Username: req.Username,
		}
		if err := rt.db.CreateUser(&newUser); err != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}
		user = &newUser
	} else if err != nil {
		ctx.Logger.WithError(err).Error("failed to get user by name")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tokenString, err := createToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := schema.LoginResponse{User: *user, Token: tokenString}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}
