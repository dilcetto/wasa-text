package api

import (
	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/dilcetto/wasa/service/database"

	"encoding/json"
	"errors"
	"fmt"
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
		fmt.Println("Generating new ID...")
		if err != nil {
			fmt.Println("Error generating ID:", err)
			http.Error(w, "Failed to generate user ID", http.StatusInternalServerError)
			return
		}
		newUser := schema.User{
			ID:       newID,
			Username: req.Username,
		}
		if err := rt.db.CreateUser(&newUser); err != nil {
			fmt.Println("Error creating user:", err)
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}
		user = &newUser
	} else if err != nil {
		fmt.Printf("ERROR: Failed to get user by name: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fmt.Println("Creating token for ID:", user.ID)
	tokenString, err := createToken(user.ID)
	if err != nil {
		fmt.Println("Token creation failed:", err)
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}
	fmt.Println("Token created successfully:", tokenString)

	response := schema.LoginResponse{User: *user, Token: tokenString}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
