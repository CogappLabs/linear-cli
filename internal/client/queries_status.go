package client

import "fmt"

// ListStatuses returns workflow states (statuses) for the given team key.
func ListStatuses(teamKey string) ([]State, error) {
	if teamKey == "" {
		return nil, fmt.Errorf("status list requires --team KEY")
	}

	q := `query($key: String!, $first: Int!) {
		workflowStates(filter: { team: { key: { eq: $key } } }, first: $first) {
			nodes {
				id
				name
				type
			}
			pageInfo { hasNextPage endCursor }
		}
	}`

	var out struct {
		WorkflowStates struct {
			Nodes    []State `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"workflowStates"`
	}
	if err := Query(q, map[string]any{"key": teamKey, "first": 100}, &out); err != nil {
		return nil, err
	}
	return out.WorkflowStates.Nodes, nil
}
