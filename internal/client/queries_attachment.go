package client

import "fmt"

// attachmentFields is the shared selection set for a single attachment.
const attachmentFields = `
	id
	title
	subtitle
	url
	createdAt
`

// Attachment is a link attached to an issue (e.g. a URL, or an
// integration-linked resource such as a GitHub PR or Slack message).
type Attachment struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
}

// ListAttachments returns the attachments on an issue (identifier or UUID).
func ListAttachments(issueID string) ([]Attachment, error) {
	q := fmt.Sprintf(`query($id: String!) {
		issue(id: $id) {
			attachments {
				nodes { %s }
			}
		}
	}`, attachmentFields)

	var out struct {
		Issue struct {
			Attachments struct {
				Nodes []Attachment `json:"nodes"`
			} `json:"attachments"`
		} `json:"issue"`
	}
	if err := Query(q, map[string]any{"id": issueID}, &out); err != nil {
		return nil, err
	}
	return out.Issue.Attachments.Nodes, nil
}

// AddAttachmentURL links a URL to an issue (issueID must be the issue UUID).
func AddAttachmentURL(issueID, url, title string) (*Attachment, error) {
	q := `mutation($issueId: String!, $url: String!, $title: String) {
		attachmentLinkURL(issueId: $issueId, url: $url, title: $title) {
			success
			attachment { id title url }
		}
	}`
	var out struct {
		AttachmentLinkURL struct {
			Success    bool       `json:"success"`
			Attachment Attachment `json:"attachment"`
		} `json:"attachmentLinkURL"`
	}
	vars := map[string]any{"issueId": issueID, "url": url}
	if title != "" {
		vars["title"] = title
	}
	if err := Query(q, vars, &out); err != nil {
		return nil, err
	}
	if !out.AttachmentLinkURL.Success {
		return nil, fmt.Errorf("attachmentLinkURL reported failure")
	}
	return &out.AttachmentLinkURL.Attachment, nil
}

// DeleteAttachment deletes an attachment by id.
func DeleteAttachment(id string) error {
	q := `mutation($id: String!) {
		attachmentDelete(id: $id) { success }
	}`
	var out struct {
		AttachmentDelete struct {
			Success bool `json:"success"`
		} `json:"attachmentDelete"`
	}
	if err := Query(q, map[string]any{"id": id}, &out); err != nil {
		return err
	}
	if !out.AttachmentDelete.Success {
		return fmt.Errorf("attachmentDelete reported failure")
	}
	return nil
}
