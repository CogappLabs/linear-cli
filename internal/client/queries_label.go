package client

import "fmt"

// labelFields is the shared selection set for a single issue label.
const labelFields = `
	id
	name
	description
	color
	isGroup
	team { id key name }
`

// LabelDetail is the full issue label shape (Label in types.go stays name-only
// for the embedded list on Issue).
type LabelDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	IsGroup     bool   `json:"isGroup"`
	Team        *Team  `json:"team"`
}

type labelConnection struct {
	Nodes    []LabelDetail `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListLabels returns every issue label. With a team key it returns the labels
// that team's issues can carry: its own plus workspace labels.
func ListLabels(teamKey string) ([]LabelDetail, error) {
	return queryLabels(labelFilter("key", teamKey))
}

// labelFilter scopes labels to one team (by "key" or "id") plus the workspace
// labels, which have no team. An empty value means no filter.
func labelFilter(field, value string) map[string]any {
	if value == "" {
		return map[string]any{}
	}
	return map[string]any{"or": []map[string]any{
		{"team": map[string]any{field: map[string]any{"eq": value}}},
		{"team": map[string]any{"null": true}},
	}}
}

func queryLabels(filter map[string]any) ([]LabelDetail, error) {
	q := fmt.Sprintf(`query($filter: IssueLabelFilter, $first: Int!, $after: String) {
		issueLabels(filter: $filter, first: $first, after: $after) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, labelFields)

	var all []LabelDetail
	vars := map[string]any{"filter": filter, "first": 250}
	for {
		var out struct {
			IssueLabels labelConnection `json:"issueLabels"`
		}
		if err := Query(q, vars, &out); err != nil {
			return nil, err
		}
		all = append(all, out.IssueLabels.Nodes...)
		if !out.IssueLabels.PageInfo.HasNextPage {
			return all, nil
		}
		vars["after"] = out.IssueLabels.PageInfo.EndCursor
	}
}

// CreateLabel creates an issue label and returns it.
func CreateLabel(input map[string]any) (*LabelDetail, error) {
	q := fmt.Sprintf(`mutation($input: IssueLabelCreateInput!) {
		issueLabelCreate(input: $input) {
			success
			issueLabel { %s }
		}
	}`, labelFields)
	var out struct {
		IssueLabelCreate struct {
			Success    bool        `json:"success"`
			IssueLabel LabelDetail `json:"issueLabel"`
		} `json:"issueLabelCreate"`
	}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.IssueLabelCreate.Success {
		return nil, fmt.Errorf("issueLabelCreate reported failure")
	}
	return &out.IssueLabelCreate.IssueLabel, nil
}

// TeamUsableLabels returns the labels an issue in the team can carry.
func TeamUsableLabels(teamID string) ([]LabelDetail, error) {
	return queryLabels(labelFilter("id", teamID))
}
