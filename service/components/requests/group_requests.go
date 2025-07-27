package requests

import (
	"net/url"
	"regexp"
)

type GroupCreateRequest struct {
	GroupName string   `json:"groupName"`
	Members   []string `json:"members"`
}

func (g *GroupCreateRequest) IsValid() bool {
	match, _ := regexp.MatchString(`^.*?$`, g.GroupName)
	return len(g.GroupName) >= 3 && len(g.GroupName) <= 50 && match && len(g.Members) >= 3 && len(g.Members) <= 100
}

type AddMemberRequest struct {
	Username string `json:"username"`
}

func (a *AddMemberRequest) IsValid() bool {
	match, _ := regexp.MatchString(`^.*?$`, a.Username)
	return len(a.Username) > 0 && match
}

type RemoveMemberRequest struct {
	Username string `json:"username"`
}

func (r *RemoveMemberRequest) IsValid() bool {
	match, _ := regexp.MatchString(`^.*?$`, r.Username)
	return match
}

type GroupNameUpdateRequest struct {
	GroupName string `json:"groupName"`
}

func (g *GroupNameUpdateRequest) IsValid() bool {
	match, _ := regexp.MatchString(`^.*?$`, g.GroupName)
	return len(g.GroupName) >= 3 && len(g.GroupName) <= 50 && match
}

type GroupPhotoUpdateRequest struct {
	PhotoURL string `json:"photo_url"`
}

func (g *GroupPhotoUpdateRequest) IsValid() bool {
	_, err := url.ParseRequestURI(g.PhotoURL)
	return err == nil
}
