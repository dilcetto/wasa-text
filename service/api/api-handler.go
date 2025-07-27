package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// profile routes
	rt.router.POST("/user/login", rt.wrap(rt.post_login))
	rt.router.PATCH("/user/username", rt.wrap(rt.patch_username))
	rt.router.PATCH("/user/photo", rt.wrap(rt.set_photo))
	// conversation routes
	rt.router.GET("/coversations", rt.wrap(rt.get_conversations))
	rt.router.GET("/conversation/:conversationId/messages", rt.wrap(rt.get_messages))
	rt.router.POST("/conversation/:conversationId/messages", rt.wrap(rt.send_messages))
	rt.router.POST("/conversations/:conversationId/messages/:messageId/forward", rt.wrap(rt.forward_message))
	rt.router.DELETE("/conversation/:conversationId/messages/:messageId", rt.wrap(rt.delete_message))
	//reaction routes
	rt.router.POST("/conversations/:conversationId/messages/:messageId/comment", rt.wrap(rt.add_reaction))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.delete_reactions))
	//group routes
	rt.router.POST("/conversations/:converstionId/group", rt.wrap(rt.create_group))
	rt.router.PATCH("/conversations/:converstionId/group/name", rt.wrap(rt.set_groupname))
	rt.router.PATCH("/conversations/:converstionId/group/photo", rt.wrap(rt.set_groupphoto))
	rt.router.POST("/conversations/:converstionId/group/members", rt.wrap(rt.add_member))
	rt.router.DELETE("/conversations/:converstionId/group/members/:username", rt.wrap(rt.leave_group))

	// special routes
	rt.router.GET("/context", rt.wrap(rt.getContextReply))
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
