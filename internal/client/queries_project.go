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

// CreateProject creates a project. input needs teamIds ([]string) and name.
func CreateProject(input map[string]any) (*Project, error) {
	q := fmt.Sprintf(`mutation($input: ProjectCreateInput!) {
		projectCreate(input: $input) { success project { %s } }
	}`, projectFields)
	var out struct {
		ProjectCreate struct {
			Success bool    `json:"success"`
			Project Project `json:"project"`
		} `json:"projectCreate"`
	}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.ProjectCreate.Success {
		return nil, fmt.Errorf("projectCreate reported failure")
	}
	return &out.ProjectCreate.Project, nil
}

// UpdateProject applies input to an existing project.
func UpdateProject(id string, input map[string]any) (*Project, error) {
	q := fmt.Sprintf(`mutation($id: String!, $input: ProjectUpdateInput!) {
		projectUpdate(id: $id, input: $input) { success project { %s } }
	}`, projectFields)
	var out struct {
		ProjectUpdate struct {
			Success bool    `json:"success"`
			Project Project `json:"project"`
		} `json:"projectUpdate"`
	}
	if err := Query(q, map[string]any{"id": id, "input": input}, &out); err != nil {
		return nil, err
	}
	if !out.ProjectUpdate.Success {
		return nil, fmt.Errorf("projectUpdate reported failure")
	}
	return &out.ProjectUpdate.Project, nil
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
