package requests

type GetConversationByIDRequest struct {
	ConversationID string `json:"conversationId" validate:"required"`
}

type GetAllConversationsRequest struct {
	// Add filters or pagination here if needed later
}
