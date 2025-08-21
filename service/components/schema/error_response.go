package schema

import "errors"

var (
	ErrUserDoesNotExist            = errors.New("user does not exist")
	ErrConversationDoesNotExist    = errors.New("conversation does not exist")
	ErrMessageDoesNotExist         = errors.New("message does not exist")
	ErrReactionDoesNotExist        = errors.New("reaction does not exist")
	ErrGroupDoesNotExist           = errors.New("group does not exist")
	ErrUnauthorizedToDeleteMessage = errors.New("unauthorized to delete message")
	ErrInvalidGroupName            = errors.New("invalid group name")
	ErrGroupNotFound               = errors.New("group not found")
)
