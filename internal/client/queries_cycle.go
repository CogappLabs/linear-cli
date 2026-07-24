package client

import "fmt"

// Cycle is a Linear iteration/sprint scoped to a team.
type Cycle struct {
	ID          string  `json:"id"`
	Number      float64 `json:"number"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartsAt    string  `json:"startsAt"`
	EndsAt      string  `json:"endsAt"`
	CompletedAt string  `json:"completedAt"`
	Team        *Team   `json:"team"`
}

type cycleConnection struct {
	Nodes    []Cycle `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListCycles returns cycles for the given team key, newest first.
func ListCycles(teamKey string) ([]Cycle, error) {
	if teamKey == "" {
		return nil, fmt.Errorf("cycle list requires --team KEY")
	}

	q := `query($key: String!, $first: Int!) {
		cycles(filter: { team: { key: { eq: $key } } }, first: $first) {
			nodes {
				id
				number
				name
				description
				startsAt
				endsAt
				completedAt
				team { id key name }
			}
			pageInfo { hasNextPage endCursor }
		}
	}`

	var out struct {
		Cycles cycleConnection `json:"cycles"`
	}
	if err := Query(q, map[string]any{"key": teamKey, "first": 50}, &out); err != nil {
		return nil, err
	}
	return out.Cycles.Nodes, nil
}
