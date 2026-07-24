package client

import "fmt"

// statusUpdateFields is the shared selection set for a single project update (status update).
const statusUpdateFields = `
	id
	body
	health
	createdAt
	user { name }
`

type StatusUpdate struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	Health    string `json:"health"`
	CreatedAt string `json:"createdAt"`
	User      *User  `json:"user"`
}

type statusUpdateConnection struct {
	Nodes    []StatusUpdate `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListStatusUpdates returns status updates (project updates) for a project.
func ListStatusUpdates(projectID string) ([]StatusUpdate, error) {
	if projectID == "" {
		return nil, fmt.Errorf("status-update list requires --project ID")
	}

	filter := map[string]any{
		"project": map[string]any{"id": map[string]any{"eq": projectID}},
	}

	q := fmt.Sprintf(`query($filter: ProjectUpdateFilter, $first: Int!) {
		projectUpdates(filter: $filter, first: $first) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, statusUpdateFields)

	var out struct {
		ProjectUpdates statusUpdateConnection `json:"projectUpdates"`
	}
	if err := Query(q, map[string]any{"filter": filter, "first": 50}, &out); err != nil {
		return nil, err
	}
	return out.ProjectUpdates.Nodes, nil
}

// CreateStatusUpdate posts a new project update.
func CreateStatusUpdate(input map[string]any) (*StatusUpdate, error) {
	q := fmt.Sprintf(`mutation($input: ProjectUpdateCreateInput!) {
		projectUpdateCreate(input: $input) {
			success
			projectUpdate { %s }
		}
	}`, statusUpdateFields)
	var out struct {
		ProjectUpdateCreate struct {
			Success       bool         `json:"success"`
			ProjectUpdate StatusUpdate `json:"projectUpdate"`
		} `json:"projectUpdateCreate"`
	}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.ProjectUpdateCreate.Success {
		return nil, fmt.Errorf("projectUpdateCreate reported failure")
	}
	return &out.ProjectUpdateCreate.ProjectUpdate, nil
}

// DeleteStatusUpdate deletes a project update by id.
func DeleteStatusUpdate(id string) error {
	q := `mutation($id: String!) {
		projectUpdateDelete(id: $id) {
			success
		}
	}`
	var out struct {
		ProjectUpdateDelete struct {
			Success bool `json:"success"`
		} `json:"projectUpdateDelete"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return err
	}
	if !out.ProjectUpdateDelete.Success {
		return fmt.Errorf("projectUpdateDelete reported failure")
	}
	return nil
}
