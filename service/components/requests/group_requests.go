package requests

type GroupCreateRequest struct {
	GroupName string   `json:"groupName"`
	Photo     []byte   `json:"photo"`
	Members   []string `json:"members"`
}

func (g *GroupCreateRequest) IsValid() bool {
	if len(g.GroupName) < 3 || len(g.GroupName) > 50 {
		return false
	}
	if len(g.Photo) == 0 {
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
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func (l *LeaveGroupRequest) IsValid() bool {
	return len(l.GroupID) > 0 && len(l.UserID) > 0
}

type SetGroupNameRequest struct {
	NewName string `json:"new_name"`
}

func (u *SetGroupNameRequest) IsValid() bool {
	return len(u.NewName) > 0
}
