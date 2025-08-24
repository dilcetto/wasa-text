package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/dilcetto/wasa/service/api/reqcontext"
	"github.com/dilcetto/wasa/service/components/requests"
	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversations); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode conversations")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Missing conversation ID", http.StatusBadRequest)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversation, err := rt.db.GetConversationByID(userID, conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	messages, err := rt.db.GetMessagesByConversationID(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to get messages")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	conversation.Messages = messages
	for _, msg := range messages {
		if err := rt.db.MarkMessageStatus(msg.ID, userID, "read"); err != nil {
			ctx.Logger.WithError(err).WithField("message_id", msg.ID).Error("Failed to mark message as read")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conversation); err != nil {
		ctx.Logger.WithError(err).Error("Failed to encode conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Missing conversation ID", http.StatusBadRequest)
		return
	}

	var message schema.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode message")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate message ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	message.ID = messageID
	message.SenderID = userID
	message.ConversationID = conversationID
	message.Timestamp = time.Now().Format(time.RFC3339)
	message.MessageStatus = "sent"

	if err := rt.db.SendMessage(&message); err != nil {
		ctx.Logger.WithError(err).Error("Failed to send message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")
	if messageID == "" {
		http.Error(w, "Missing message ID", http.StatusBadRequest)
		return
	}

	var req requests.ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode forward message request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch the original message
	originalMessage, err := rt.db.GetMessageByID(messageID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to fetch original message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate new message ID
	newMessageID, err := generateNewID()
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to generate new message ID")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create the forwarded message
	forwardedMessage := schema.Message{
		ID:             newMessageID,
		SenderID:       userID,
		Sender:         originalMessage.Sender,
		ConversationID: req.TargetConversationID,
		MessageType:    originalMessage.MessageType,
		Content:        originalMessage.Content,
		Timestamp:      time.Now().Format(time.RFC3339),
		MessageStatus:  "sent",
		Reaction:       []schema.Reaction{},
		Attachments:    originalMessage.Attachments,
		ForwardedFrom:  originalMessage.SenderID,
	}

	if err := rt.db.SendMessage(&forwardedMessage); err != nil {
		ctx.Logger.WithError(err).Error("Failed to send forwarded message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = rt.db.DeleteMessage(conversationID, messageID, userID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Failed to delete message")
		ctx.Logger.WithFields(logrus.Fields{
			"conversation_id": conversationID,
			"message_id":      messageID,
			"user_id":         userID,
		}).Error("Failed to delete message")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	ctx.Logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"message_id":      messageID,
		"user_id":         userID,
	}).Info("Message deleted successfully")
}

func (rt *_router) setMessageStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	if conversationID == "" || messageID == "" {
		http.Error(w, "Missing conversation or message ID", http.StatusBadRequest)
		return
	}

	userID, err := rt.getAuthenticatedUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Error("Failed to decode message status request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := rt.db.MarkMessageStatus(messageID, userID, req.Status); err != nil {
		ctx.Logger.WithError(err).Error("Failed to update message status")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
