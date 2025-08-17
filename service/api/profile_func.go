package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
)

// post_login handles user login requests
func (rt *_router) post_login(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var request requests.Username
	err := json.NewDecoder(r.Body).Decode(&request)
	_ = r.Body.Close()

	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode login request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !request.IsValid() {
		ctx.Logger.WithField("Username", request.Username).Error("Invalid username provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid, created, err := rt.db.insert_user(request.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to insert user into database")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Logger.Debug("post_login: User inserted successfully")
	result := schema.LoginResponse{ID: uid}
	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(result)
}

// patch_username handles requests to change the username of a user
func (rt *_router) patch_username(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var request requests.Username
	err := json.NewDecoder(r.Body).Decode(&request)
	_ = r.Body.Close()

	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode username change request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !request.IsValid() {
		ctx.Logger.WithField("Username", request.Username).Error("Invalid username provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.update_username(ctx.Uid, request.Username)
	if errors.Is(err, schema.ErrUsernameAlreadyExists) {
		ctx.Logger.WithError(err).Error("Failed to update username in database")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Logger.Debug("patch_username: Username updated successfully")
	result := map[string]string{
		"message":  "Username updated successfully",
		"username": request.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// set_photo handles requests to change the user's profile photo
func (rt *_router) set_photo(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var request requests.Photo
	err := json.NewDecoder(r.Body).Decode(&request)
	_ = r.Body.Close()

	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode photo change request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !request.IsValid() {
		ctx.Logger.Error("Invalid photo provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.update_photo(ctx.Uid, request.Photo)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to update photo in database")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Logger.Debug("set_photo: Photo updated successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"photoUrl": request.Photo,
	})
}
