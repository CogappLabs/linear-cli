package client

import "fmt"

// documentListFields is the selection set for a document in list results.
// Excludes content, which can be large; fetched only by GetDocument.
const documentListFields = `
	id
	title
	url
	icon
	createdAt
	updatedAt
	creator { name }
	project { id name }
`

// documentFields is the selection set for a single document, including content.
const documentFields = `
	id
	title
	content
	url
	icon
	createdAt
	updatedAt
	creator { name }
	project { id name }
`

type Document struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	URL       string   `json:"url"`
	Icon      string   `json:"icon"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	Creator   *User    `json:"creator"`
	Project   *Project `json:"project"`
}

type documentConnection struct {
	Nodes    []Document `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

// ListDocuments returns all documents accessible to the user, newest updated first.
func ListDocuments() ([]Document, error) {
	q := fmt.Sprintf(`query($first: Int!) {
		documents(first: $first, orderBy: updatedAt) {
			nodes { %s }
			pageInfo { hasNextPage endCursor }
		}
	}`, documentListFields)

	var out struct {
		Documents documentConnection `json:"documents"`
	}
	if err := Query(q, map[string]any{"first": 50}, &out); err != nil {
		return nil, err
	}
	return out.Documents.Nodes, nil
}

// CreateDocument creates a document. input needs title, plus teamId or projectId.
func CreateDocument(input map[string]any) (*Document, error) {
	q := fmt.Sprintf(`mutation($input: DocumentCreateInput!) {
		documentCreate(input: $input) { success document { %s } }
	}`, documentFields)
	var out struct {
		DocumentCreate struct {
			Success  bool     `json:"success"`
			Document Document `json:"document"`
		} `json:"documentCreate"`
	}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.DocumentCreate.Success {
		return nil, fmt.Errorf("documentCreate reported failure")
	}
	return &out.DocumentCreate.Document, nil
}

// UpdateDocument applies input to an existing document.
func UpdateDocument(id string, input map[string]any) (*Document, error) {
	q := fmt.Sprintf(`mutation($id: String!, $input: DocumentUpdateInput!) {
		documentUpdate(id: $id, input: $input) { success document { %s } }
	}`, documentFields)
	var out struct {
		DocumentUpdate struct {
			Success  bool     `json:"success"`
			Document Document `json:"document"`
		} `json:"documentUpdate"`
	}
	if err := Query(q, map[string]any{"id": id, "input": input}, &out); err != nil {
		return nil, err
	}
	if !out.DocumentUpdate.Success {
		return nil, fmt.Errorf("documentUpdate reported failure")
	}
	return &out.DocumentUpdate.Document, nil
}

// GetDocument fetches one document by id or slug id, including its content.
func GetDocument(id string) (*Document, error) {
	q := fmt.Sprintf(`query($id: String!) {
		document(id: $id) { %s }
	}`, documentFields)
	var out struct {
		Document Document `json:"document"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return nil, err
	}
	return &out.Document, nil
}
