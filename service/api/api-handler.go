package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	//auth
	rt.router.POST("/login", rt.wrap(rt.doLogin))
	// profile routes
	rt.router.GET("/searchby", rt.wrap(rt.search_by))
	rt.router.PUT("/user/username", rt.wrap(rt.setMyUserName))
	rt.router.PUT("/user/photo", rt.wrap(rt.setMyPhoto))
	// conversation and messages routes
	rt.router.GET("/coversations", rt.wrap(rt.getMyConversations))
	rt.router.GET("/conversation/:conversationId", rt.wrap(rt.getConversation))
	rt.router.POST("/conversation/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.POST("/conversations/:conversationId/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.DELETE("/conversation/:conversationId/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/conversations/:conversationId/messages/:messageId/status", rt.wrap(rt.setMessageStatus))
	//reaction routes
	rt.router.POST("/conversations/:conversationId/messages/:messageId/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/comment", rt.wrap(rt.uncommentMessage))
	//group routes
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.POST("/groups/:groupId", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
