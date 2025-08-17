package requests

import (
	"net/url"
)

type GroupCreateRequest struct {
	GroupName string   `json:"groupName"`
	PhotoURL  string   `json:"photo_url"`
	Members   []string `json:"members"`
}

func (g *GroupCreateRequest) IsValid() bool {
	if len(g.GroupName) < 3 || len(g.GroupName) > 50 {
		return false
	}
	if _, err := url.ParseRequestURI(g.PhotoURL); err != nil {
		return false
	}
	for _, member := range g.Members {
		if len(member) == 0 {
			return false
		}
	}
	return true
}

type AddMemberRequest struct {
	Username string `json:"username"`
	GroupID  string `json:"group_id"`
}

func (a *AddMemberRequest) IsValid() bool {
	return len(a.Username) > 0 && len(a.GroupID) > 0
}

type LeaveGroupRequest struct {
	GroupID string `json:"group_id"`
}

func (l *LeaveGroupRequest) IsValid() bool {
	return len(l.GroupID) > 0
}
