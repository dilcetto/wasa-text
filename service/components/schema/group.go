package schema

type Group struct {
	GroupName string   `json:"group_name"`
	Members   []string `json:"members"`
}
