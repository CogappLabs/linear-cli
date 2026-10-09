package client

import (
	"fmt"
	"strings"
)

const relationFields = `id type issue { id identifier title } relatedIssue { id identifier title }`

// CreateRelation records "issueID <relType> relatedID" (both UUIDs or identifiers).
func CreateRelation(issueID, relatedID, relType string) error {
	q := `mutation($input: IssueRelationCreateInput!) {
		issueRelationCreate(input: $input) { success }
	}`
	var out struct {
		IssueRelationCreate struct {
			Success bool `json:"success"`
		} `json:"issueRelationCreate"`
	}
	input := map[string]any{"issueId": issueID, "relatedIssueId": relatedID, "type": relType}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return err
	}
	if !out.IssueRelationCreate.Success {
		return fmt.Errorf("issueRelationCreate reported failure")
	}
	return nil
}

// DeleteRelation removes a relation by its id.
func DeleteRelation(id string) error {
	q := `mutation($id: String!) { issueRelationDelete(id: $id) { success } }`
	var out struct {
		IssueRelationDelete struct {
			Success bool `json:"success"`
		} `json:"issueRelationDelete"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return err
	}
	if !out.IssueRelationDelete.Success {
		return fmt.Errorf("issueRelationDelete reported failure")
	}
	return nil
}

// RelationIDsBetween returns the ids of every relation linking is to the issue
// with the given identifier, in either direction.
func RelationIDsBetween(is *Issue, otherIdentifier string) []string {
	var ids []string
	for _, r := range is.Relations.Nodes {
		if r.RelatedIssue != nil && strings.EqualFold(r.RelatedIssue.Identifier, otherIdentifier) {
			ids = append(ids, r.ID)
		}
	}
	for _, r := range is.InverseRelations.Nodes {
		if r.Issue != nil && strings.EqualFold(r.Issue.Identifier, otherIdentifier) {
			ids = append(ids, r.ID)
		}
	}
	return ids
}
