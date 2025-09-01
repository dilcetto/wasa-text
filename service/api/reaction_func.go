package api

import (
	"encoding/json"
	"net/http"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req requests.AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode comment request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	messageID := ps.ByName("messageId")
	if messageID == "" {
		http.Error(w, "Missing message ID", http.StatusBadRequest)
		return
	}

	reaction := &schema.Reaction{
		MessageId: messageID,
		UserId:    userID,
		Emoji:     req.Emoji,
	}

	err = rt.db.AddReactionToMessage(reaction)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to insert comment")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req requests.RemoveReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode un-comment request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	messageID := ps.ByName("messageId")
	if messageID == "" {
		http.Error(w, "Missing message ID", http.StatusBadRequest)
		return
	}

	err = rt.db.DeleteReactionFromMessage(messageID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to remove comment")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
