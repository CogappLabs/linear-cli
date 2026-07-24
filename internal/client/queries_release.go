package client

import "fmt"

// releaseListFields is the selection set for a release in list results.
const releaseListFields = `
	id
	name
	version
	slugId
	createdAt
	updatedAt
	completedAt
	stage { id name }
	pipeline { id name }
`

// releaseFields is the selection set for a single release.
const releaseFields = `
	id
	name
	version
	slugId
	description
	url
	createdAt
	updatedAt
	completedAt
	canceledAt
	startedAt
	commitSha
	stage { id name }
	pipeline { id name }
	creator { name }
`

type Release struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	SlugID      string           `json:"slugId"`
	Description string           `json:"description"`
	URL         string           `json:"url"`
	CreatedAt   string           `json:"createdAt"`
	UpdatedAt   string           `json:"updatedAt"`
	CompletedAt string           `json:"completedAt"`
	CanceledAt  string           `json:"canceledAt"`
	StartedAt   string           `json:"startedAt"`
	CommitSha   string           `json:"commitSha"`
	Stage       *ReleaseStage    `json:"stage"`
	Pipeline    *ReleasePipeline `json:"pipeline"`
	Creator     *User            `json:"creator"`
}

type ReleaseStage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ReleasePipeline struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type releaseConnection struct {
	Nodes    []Release `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListReleases returns up to 50 releases in the workspace.
func ListReleases() ([]Release, error) {
	q := fmt.Sprintf(`query($first: Int!) {
		releases(first: $first) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, releaseListFields)

	var out struct {
		Releases releaseConnection `json:"releases"`
	}
	if err := Query(q, map[string]any{"first": 50}, &out); err != nil {
		return nil, err
	}
	return out.Releases.Nodes, nil
}

// GetRelease fetches one release by its UUID or slug id.
func GetRelease(id string) (*Release, error) {
	q := fmt.Sprintf(`query($id: String!) {
		release(id: $id) { %s }
	}`, releaseFields)
	var out struct {
		Release Release `json:"release"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.Release, nil
}

// releaseNoteFields is the selection set for a release note.
const releaseNoteFields = `
	id
	title
	slugId
	url
	createdAt
	updatedAt
	releaseCount
`

type ReleaseNote struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	SlugID       string `json:"slugId"`
	URL          string `json:"url"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	ReleaseCount int    `json:"releaseCount"`
}

type releaseNoteConnection struct {
	Nodes    []ReleaseNote `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListReleaseNotes returns up to 50 release notes that cover the given release,
// filtered via ReleaseNoteFilter.release.id (ReleaseFilter has an id comparator).
func ListReleaseNotes(releaseID string) ([]ReleaseNote, error) {
	q := fmt.Sprintf(`query($first: Int!, $releaseId: ID!) {
		releaseNotes(first: $first, filter: { release: { id: { eq: $releaseId } } }) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, releaseNoteFields)

	var out struct {
		ReleaseNotes releaseNoteConnection `json:"releaseNotes"`
	}
	if err := Query(q, map[string]any{"first": 50, "releaseId": releaseID}, &out); err != nil {
		return nil, err
	}
	return out.ReleaseNotes.Nodes, nil
}
