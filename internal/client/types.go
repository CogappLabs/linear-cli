package client

// Shared response shapes for Linear GraphQL. Fields track only what the CLI
// renders; add to these structs as commands need more.

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Active      bool   `json:"active"`
}

type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Team struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Label struct {
	Name string `json:"name"`
}

type Comment struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	User      *User  `json:"user"`
}

type Issue struct {
	ID          string `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    int    `json:"priority"`
	BranchName  string `json:"branchName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	State       *State `json:"state"`
	Assignee    *User  `json:"assignee"`
	Team        *Team  `json:"team"`
	Labels      struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	Comments struct {
		Nodes []Comment `json:"nodes"`
	} `json:"comments"`
	Relations struct {
		Nodes []Relation `json:"nodes"`
	} `json:"relations"`
	InverseRelations struct {
		Nodes []Relation `json:"nodes"`
	} `json:"inverseRelations"`
}

// Relation is one IssueRelation edge: Issue <Type> RelatedIssue
// (e.g. Issue blocks RelatedIssue).
type Relation struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Issue        *RelIssue `json:"issue"`
	RelatedIssue *RelIssue `json:"relatedIssue"`
}

type RelIssue struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
}

type issueConnection struct {
	Nodes    []Issue `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}
