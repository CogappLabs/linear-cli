package client

import "fmt"

// projectFields is the shared selection set for a single project.
const projectFields = `
	id
	name
	description
	url
	slugId
	progress
	startDate
	targetDate
	status { name }
	lead { id name displayName email }
`

type Project struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	SlugID      string  `json:"slugId"`
	Progress    float64 `json:"progress"`
	StartDate   string  `json:"startDate"`
	TargetDate  string  `json:"targetDate"`
	Status      *Status `json:"status"`
	Lead        *User   `json:"lead"`
}

type Status struct {
	Name string `json:"name"`
}

type projectConnection struct {
	Nodes    []Project `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListProjects returns projects, optionally filtered to those accessible by a
// team, newest updated first.
func ListProjects(teamKey string) ([]Project, error) {
	filter := map[string]any{}
	if teamKey != "" {
		filter["accessibleTeams"] = map[string]any{
			"some": map[string]any{"key": map[string]any{"eq": teamKey}},
		}
	}

	q := fmt.Sprintf(`query($filter: ProjectFilter, $first: Int!) {
		projects(filter: $filter, first: $first, orderBy: updatedAt) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, projectFields)

	var out struct {
		Projects projectConnection `json:"projects"`
	}
	if err := Query(q, map[string]any{"filter": filter, "first": 50}, &out); err != nil {
		return nil, err
	}
	return out.Projects.Nodes, nil
}

// GetProject fetches one project by id or slug id.
func GetProject(id string) (*Project, error) {
	q := fmt.Sprintf(`query($id: String!) {
		project(id: $id) { %s }
	}`, projectFields)
	var out struct {
		Project Project `json:"project"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.Project, nil
}
