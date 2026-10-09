package client

import (
	"fmt"
	"strings"
)

// issueFields is the shared selection set for a single issue.
const issueFields = `
	id
	identifier
	title
	description
	url
	priority
	branchName
	createdAt
	updatedAt
	state { id name type }
	assignee { id name displayName email }
	team { id key name }
	labels { nodes { name } }
`

// Viewer returns the authenticated user.
func Viewer() (*User, error) {
	var out struct {
		Viewer User `json:"viewer"`
	}
	err := Query(`query { viewer { id name displayName email } }`, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out.Viewer, nil
}

// GetIssue fetches one issue by its identifier (e.g. ABC-123) or UUID,
// including comments.
func GetIssue(id string) (*Issue, error) {
	q := fmt.Sprintf(`query($id: String!) {
		issue(id: $id) {
			%s
			comments { nodes { id body createdAt user { name displayName } } }
			relations(first: 100) { nodes { %[2]s } }
			inverseRelations(first: 100) { nodes { %[2]s } }
		}
	}`, issueFields, relationFields)
	var out struct {
		Issue Issue `json:"issue"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.Issue, nil
}

// IssueFilter carries optional list filters. Empty fields are omitted.
type IssueFilter struct {
	TeamKey    string
	AssigneeID string   // "me" resolved by caller
	StateTypes []string // state types, or state names (case-insensitive) for anything that is not a type
	NotDone    bool     // exclude completed/canceled (overrides StateTypes)
	First      int      // page size / cap; ignored when All is set
	All        bool     // page through every match
}

// buildIssueFilter renders an IssueFilter into the GraphQL filter object.
func buildIssueFilter(f IssueFilter) map[string]any {
	filter := map[string]any{}
	if f.TeamKey != "" {
		filter["team"] = map[string]any{"key": map[string]any{"eq": f.TeamKey}}
	}
	if f.AssigneeID != "" {
		filter["assignee"] = map[string]any{"id": map[string]any{"eq": f.AssigneeID}}
	}
	switch {
	case f.NotDone:
		filter["state"] = map[string]any{"type": map[string]any{"nin": []string{"completed", "canceled"}}}
	case len(f.StateTypes) > 0:
		filter["state"] = buildStateFilter(f.StateTypes)
	}
	return filter
}

var stateTypes = map[string]bool{
	"triage": true, "backlog": true, "unstarted": true, "started": true,
	"completed": true, "canceled": true, "duplicate": true,
}

// buildStateFilter treats each value as a state type when it is one, and as a
// workflow state name otherwise. Teams name their states freely ("Review"),
// and those names are not types, so a type-only filter matched nothing.
func buildStateFilter(values []string) map[string]any {
	var types []string
	var clauses []any
	for _, v := range values {
		if stateTypes[strings.ToLower(v)] {
			types = append(types, strings.ToLower(v))
		} else {
			clauses = append(clauses, map[string]any{"name": map[string]any{"eqIgnoreCase": v}})
		}
	}
	if len(clauses) == 0 {
		return map[string]any{"type": map[string]any{"in": types}}
	}
	if len(types) > 0 {
		clauses = append(clauses, map[string]any{"type": map[string]any{"in": types}})
	}
	return map[string]any{"or": clauses}
}

// ListIssues returns issues matching the filter, newest updated first. When
// f.All is set it pages through the entire result; otherwise it returns a
// single page of f.First (default 50).
func ListIssues(f IssueFilter) ([]Issue, error) {
	filter := buildIssueFilter(f)

	q := fmt.Sprintf(`query($filter: IssueFilter, $first: Int!, $after: String) {
		issues(filter: $filter, first: $first, after: $after, orderBy: updatedAt) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, issueFields)

	pageSize := f.First
	if pageSize <= 0 || (f.All && pageSize > 250) {
		pageSize = 50
	}
	if f.All {
		pageSize = 250 // fewer round-trips for a full pull
	}

	all := []Issue{} // non-nil so --json prints [] rather than null
	var after string
	for {
		var out struct {
			Issues issueConnection `json:"issues"`
		}
		vars := map[string]any{"filter": filter, "first": pageSize}
		if after != "" {
			vars["after"] = after
		}
		if err := Query(q, vars, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Issues.Nodes...)
		if !f.All || !out.Issues.PageInfo.HasNextPage {
			return all, nil
		}
		after = out.Issues.PageInfo.EndCursor
	}
}

// ListComments returns every comment on an issue (identifier or UUID), paging
// through the connection. GetIssue only carries the first page for a glance;
// this is the thorough fetch for `issue comments`.
func ListComments(issueID string) ([]Comment, error) {
	q := `query($id: String!, $first: Int!, $after: String) {
		issue(id: $id) {
			comments(first: $first, after: $after) {
				nodes { id body createdAt user { name displayName } }
				pageInfo { hasNextPage endCursor }
			}
		}
	}`
	var all []Comment
	var after string
	for {
		var out struct {
			Issue struct {
				Comments struct {
					Nodes    []Comment `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"comments"`
			} `json:"issue"`
		}
		vars := map[string]any{"id": issueID, "first": 100}
		if after != "" {
			vars["after"] = after
		}
		if err := Query(q, vars, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Issue.Comments.Nodes...)
		if !out.Issue.Comments.PageInfo.HasNextPage {
			return all, nil
		}
		after = out.Issue.Comments.PageInfo.EndCursor
	}
}

// CreateIssue creates an issue and returns it.
func CreateIssue(input map[string]any) (*Issue, error) {
	q := fmt.Sprintf(`mutation($input: IssueCreateInput!) {
		issueCreate(input: $input) {
			success
			issue { %s }
		}
	}`, issueFields)
	var out struct {
		IssueCreate struct {
			Success bool  `json:"success"`
			Issue   Issue `json:"issue"`
		} `json:"issueCreate"`
	}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.IssueCreate.Success {
		return nil, fmt.Errorf("issueCreate reported failure")
	}
	return &out.IssueCreate.Issue, nil
}

// UpdateIssue applies input to the issue identified by id (identifier or UUID).
func UpdateIssue(id string, input map[string]any) (*Issue, error) {
	q := fmt.Sprintf(`mutation($id: String!, $input: IssueUpdateInput!) {
		issueUpdate(id: $id, input: $input) {
			success
			issue { %s }
		}
	}`, issueFields)
	var out struct {
		IssueUpdate struct {
			Success bool  `json:"success"`
			Issue   Issue `json:"issue"`
		} `json:"issueUpdate"`
	}
	if err := Query(q, map[string]any{"id": id, "input": input}, &out); err != nil {
		return nil, err
	}
	if !out.IssueUpdate.Success {
		return nil, fmt.Errorf("issueUpdate reported failure")
	}
	return &out.IssueUpdate.Issue, nil
}

// AddComment posts a comment on an issue (UUID required by the API).
func AddComment(issueID, body string) error {
	q := `mutation($input: CommentCreateInput!) {
		commentCreate(input: $input) { success }
	}`
	var out struct {
		CommentCreate struct {
			Success bool `json:"success"`
		} `json:"commentCreate"`
	}
	input := map[string]any{"issueId": issueID, "body": body}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return err
	}
	if !out.CommentCreate.Success {
		return fmt.Errorf("commentCreate reported failure")
	}
	return nil
}

// EditComment replaces the body of a comment (by comment UUID).
func EditComment(commentID, body string) error {
	q := `mutation($id: String!, $input: CommentUpdateInput!) {
		commentUpdate(id: $id, input: $input) { success }
	}`
	var out struct {
		CommentUpdate struct {
			Success bool `json:"success"`
		} `json:"commentUpdate"`
	}
	vars := map[string]any{"id": commentID, "input": map[string]any{"body": body}}
	if err := Query(q, vars, &out); err != nil {
		return err
	}
	if !out.CommentUpdate.Success {
		return fmt.Errorf("commentUpdate reported failure")
	}
	return nil
}

// DeleteComment removes a comment (by comment UUID).
func DeleteComment(commentID string) error {
	q := `mutation($id: String!) { commentDelete(id: $id) { success } }`
	var out struct {
		CommentDelete struct {
			Success bool `json:"success"`
		} `json:"commentDelete"`
	}
	if err := Query(q, map[string]any{"id": commentID}, &out); err != nil {
		return err
	}
	if !out.CommentDelete.Success {
		return fmt.Errorf("commentDelete reported failure")
	}
	return nil
}

// ListTeams returns all teams.
func ListTeams() ([]Team, error) {
	var out struct {
		Teams struct {
			Nodes []Team `json:"nodes"`
		} `json:"teams"`
	}
	err := Query(`query { teams { nodes { id key name } } }`, nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Teams.Nodes, nil
}

// ListUsers returns all workspace members, paging through the connection
// (the API caps a page at 250, and workspaces routinely exceed one page).
func ListUsers() ([]User, error) {
	q := `query($after: String) {
		users(first: 250, after: $after) {
			nodes { id name displayName email active }
			pageInfo { hasNextPage endCursor }
		}
	}`
	var all []User
	var after string
	for {
		var out struct {
			Users struct {
				Nodes    []User `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"users"`
		}
		vars := map[string]any{}
		if after != "" {
			vars["after"] = after
		}
		if err := Query(q, vars, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Users.Nodes...)
		if !out.Users.PageInfo.HasNextPage {
			return all, nil
		}
		after = out.Users.PageInfo.EndCursor
	}
}

// GetTeam resolves a team by key and returns core fields plus a description.
func GetTeam(key string) (*Team, error) {
	// The team key is not a valid `team(id:)` argument, so filter by key.
	q := `query($key: String!) {
		teams(filter: { key: { eq: $key } }, first: 1) {
			nodes { id key name }
		}
	}`
	var out struct {
		Teams struct {
			Nodes []Team `json:"nodes"`
		} `json:"teams"`
	}
	if err := Query(q, map[string]any{"key": key}, &out); err != nil {
		return nil, err
	}
	if len(out.Teams.Nodes) == 0 {
		return nil, fmt.Errorf("no team with key %q", key)
	}
	return &out.Teams.Nodes[0], nil
}

// GetUser resolves a workspace member by email.
func GetUser(email string) (*User, error) {
	users, err := ListUsers()
	if err != nil {
		return nil, err
	}
	for i := range users {
		if strings.EqualFold(users[i].Email, email) {
			return &users[i], nil
		}
	}
	return nil, fmt.Errorf("no user with email %q", email)
}

// TeamStates returns the workflow states for a team, keyed by team UUID.
func TeamStates(teamID string) ([]State, error) {
	q := `query($id: String!) {
		team(id: $id) { states { nodes { id name type position } } }
	}`
	var out struct {
		Team struct {
			States struct {
				Nodes []State `json:"nodes"`
			} `json:"states"`
		} `json:"team"`
	}
	if err := Query(q, map[string]any{"id": teamID}, &out); err != nil {
		return nil, err
	}
	return out.Team.States.Nodes, nil
}

// TeamByKey resolves a team key (e.g. "ENG") to a team, or nil if absent.
func TeamByKey(key string) (*Team, error) {
	teams, err := ListTeams()
	if err != nil {
		return nil, err
	}
	for i := range teams {
		if teams[i].Key == key {
			return &teams[i], nil
		}
	}
	return nil, fmt.Errorf("no team with key %q", key)
}
