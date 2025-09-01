package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// require authentication and include creator among members
	userID, authErr := rt.getAuthenticatedUserID(r)
	if authErr != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		GroupName  string   `json:"groupName"`
		Members    []string `json:"members"`
		GroupPhoto []byte   `json:"groupPhoto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	groupName := body.GroupName
	members := body.Members
	photo := body.GroupPhoto

	// validate basic input
	if groupName == "" || len(members) == 0 {
		http.Error(w, "Group name and members are required", http.StatusBadRequest)
		return
	}

	// ensure creator is included
	foundCreator := false
	for _, m := range members {
		if m == userID {
			foundCreator = true
			break
		}
	}
	if !foundCreator {
		members = append(members, userID)
	}

	groupID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate group ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	createdAt := generateCurrentTimestamp()

	group := &schema.Group{
		ID:         groupID,
		GroupName:  groupName,
		GroupPhoto: photo,
		Members:    members,
		CreatedAt:  createdAt,
	}

	if err := rt.db.CreateGroup(group); err != nil {
		ctx.Logger.WithError(err).Error("Failed to create group in database")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// return the created conversation as response
	conv, err := rt.db.GetConversationByID(userID, groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Group created but failed to load conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(conv); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode group response")
	}
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req requests.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// map username to user ID
	u, err := rt.db.GetUserByName(req.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("User not found for addToGroup")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := rt.db.AddUserToGroup(groupID, u.ID); err != nil {
		ctx.Logger.WithError(err).Error("Failed to add user to group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if _, err := rt.getAuthenticatedUserID(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req requests.LeaveGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := rt.db.LeaveGroup(groupID, req.UserID); err != nil {
		ctx.Logger.WithError(err).Error("Failed to remove user from group")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	groupID := ps.ByName("groupId")
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.NewName); err != nil {
		if errors.Is(err, schema.ErrGroupNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		} else if errors.Is(err, schema.ErrInvalidGroupName) {
			http.Error(w, "Invalid group name", http.StatusBadRequest)
			return
		}
		ctx.Logger.WithError(err).Error("Failed to update group name")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// return updated conversation
	conv, err := rt.db.GetConversationByID(userID, groupID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Group name updated but failed to load conversation")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(conv)
}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	_, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		GroupPhoto []byte `json:"groupPhoto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	photo := body.GroupPhoto
	if len(photo) == 0 {
		http.Error(w, "Missing group photo", http.StatusBadRequest)
		return
	}

	if len(photo) > 10*1024*1024 {
		http.Error(w, "Photo too large. Maximum allowed size is 10 MB.", http.StatusRequestEntityTooLarge)
		return
	}

	fileType := http.DetectContentType(photo)
	if fileType != "image/jpeg" && fileType != "image/png" {
		http.Error(w, "Invalid file type. Only JPEG and PNG are supported.", http.StatusUnsupportedMediaType)
		return
	}
	if err := rt.db.UpdateGroupPhoto(groupID, photo); err != nil {
		ctx.Logger.WithError(err).Error("Failed to update group photo")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Message    string `json:"message"`
		GroupPhoto []byte `json:"groupPhoto"`
	}{
		Message:    "Photo updated successfully",
		GroupPhoto: photo,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode photo update response")
	}
}
