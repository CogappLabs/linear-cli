package client

import "fmt"

// agentSkillListFields is the selection set for an agent skill in list results.
// Excludes body, which can be large; fetched only by GetAgentSkill.
const agentSkillListFields = `
	id
	title
	description
	slugId
	shared
	createdAt
	updatedAt
	creator { name }
	owner { name }
`

// agentSkillFields is the selection set for a single agent skill, including body.
const agentSkillFields = `
	id
	title
	description
	body
	slugId
	shared
	createdAt
	updatedAt
	creator { name }
	owner { name }
`

type AgentSkill struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
	SlugID      string `json:"slugId"`
	Shared      bool   `json:"shared"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Creator     *User  `json:"creator"`
	Owner       *User  `json:"owner"`
}

type agentSkillConnection struct {
	Nodes    []AgentSkill `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListAgentSkills returns all agent skills accessible to the user, newest updated first.
func ListAgentSkills() ([]AgentSkill, error) {
	q := fmt.Sprintf(`query($first: Int!) {
		agentSkills(first: $first, orderBy: updatedAt) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, agentSkillListFields)

	var out struct {
		AgentSkills agentSkillConnection `json:"agentSkills"`
	}
	if err := Query(q, map[string]any{"first": 50}, &out); err != nil {
		return nil, err
	}
	return out.AgentSkills.Nodes, nil
}

// GetAgentSkill fetches one agent skill by id, including its body.
func GetAgentSkill(id string) (*AgentSkill, error) {
	q := fmt.Sprintf(`query($id: String!) {
		agentSkill(id: $id) { %s }
	}`, agentSkillFields)
	var out struct {
		AgentSkill AgentSkill `json:"agentSkill"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.AgentSkill, nil
}
