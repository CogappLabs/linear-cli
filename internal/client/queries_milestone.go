package client

import "fmt"

// milestoneFields is the shared selection set for a single project milestone.
const milestoneFields = `
	id
	name
	description
	targetDate
	progress
	project { id name }
`

type Milestone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TargetDate  string   `json:"targetDate"`
	Progress    float64  `json:"progress"`
	Project     *Project `json:"project"`
}

type milestoneConnection struct {
	Nodes    []Milestone `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListMilestones returns milestones for a project.
func ListMilestones(projectID string) ([]Milestone, error) {
	if projectID == "" {
		return nil, fmt.Errorf("milestone list requires --project ID")
	}

	filter := map[string]any{
		"project": map[string]any{"id": map[string]any{"eq": projectID}},
	}

	q := fmt.Sprintf(`query($filter: ProjectMilestoneFilter, $first: Int!) {
		projectMilestones(filter: $filter, first: $first) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, milestoneFields)

	var out struct {
		ProjectMilestones milestoneConnection `json:"projectMilestones"`
	}
	if err := Query(q, map[string]any{"filter": filter, "first": 50}, &out); err != nil {
		return nil, err
	}
	return out.ProjectMilestones.Nodes, nil
}

// GetMilestone fetches one project milestone by id.
func GetMilestone(id string) (*Milestone, error) {
	q := fmt.Sprintf(`query($id: String!) {
		projectMilestone(id: $id) { %s }
	}`, milestoneFields)
	var out struct {
		ProjectMilestone Milestone `json:"projectMilestone"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.ProjectMilestone, nil
}
