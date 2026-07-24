package client

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

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

// UploadAndAttach uploads a local file to Linear's storage and links it to an
// issue. Three steps: fileUpload (get a signed URL + headers), PUT the bytes,
// then attachmentCreate with the resulting asset URL. issueID must be the UUID.
func UploadAndAttach(issueID, path, title string) (*Attachment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(path)
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if title == "" {
		title = filename
	}

	// Step 1: request a signed upload URL.
	uf, err := requestUpload(contentType, filename, len(data))
	if err != nil {
		return nil, err
	}

	// Step 2: PUT the raw bytes to the signed URL with the required headers.
	if err := putBytes(uf.UploadURL, contentType, uf.Headers, data); err != nil {
		return nil, err
	}

	// Step 3: link the uploaded asset to the issue.
	return createAttachment(issueID, uf.AssetURL, title)
}

type uploadHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type uploadFile struct {
	UploadURL string         `json:"uploadUrl"`
	AssetURL  string         `json:"assetUrl"`
	Headers   []uploadHeader `json:"headers"`
}

func requestUpload(contentType, filename string, size int) (*uploadFile, error) {
	q := `mutation($contentType: String!, $filename: String!, $size: Int!) {
		fileUpload(contentType: $contentType, filename: $filename, size: $size) {
			success
			uploadFile { uploadUrl assetUrl headers { key value } }
		}
	}`
	var out struct {
		FileUpload struct {
			Success    bool        `json:"success"`
			UploadFile *uploadFile `json:"uploadFile"`
		} `json:"fileUpload"`
	}
	vars := map[string]any{"contentType": contentType, "filename": filename, "size": size}
	if err := Query(q, vars, &out); err != nil {
		return nil, err
	}
	if !out.FileUpload.Success || out.FileUpload.UploadFile == nil {
		return nil, fmt.Errorf("fileUpload did not return an upload URL")
	}
	return out.FileUpload.UploadFile, nil
}

func putBytes(url, contentType string, headers []uploadHeader, data []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	// Cache-Control and any signature headers Linear requires come back with the
	// signed URL and must be replayed verbatim.
	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("upload PUT failed: %d %s\n%s", res.StatusCode, res.Status, string(body))
	}
	return nil
}

func createAttachment(issueID, assetURL, title string) (*Attachment, error) {
	q := fmt.Sprintf(`mutation($input: AttachmentCreateInput!) {
		attachmentCreate(input: $input) { success attachment { %s } }
	}`, attachmentFields)
	var out struct {
		AttachmentCreate struct {
			Success    bool       `json:"success"`
			Attachment Attachment `json:"attachment"`
		} `json:"attachmentCreate"`
	}
	input := map[string]any{"issueId": issueID, "url": assetURL, "title": title}
	if err := Query(q, map[string]any{"input": input}, &out); err != nil {
		return nil, err
	}
	if !out.AttachmentCreate.Success {
		return nil, fmt.Errorf("attachmentCreate reported failure")
	}
	return &out.AttachmentCreate.Attachment, nil
}

// markdownImageRe matches image URLs in an issue's markdown: ![alt](url).
var markdownImageRe = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)

// ExtractImages returns image URLs embedded in an issue's markdown description
// (issue identifier or UUID).
func ExtractImages(issueID string) ([]string, error) {
	is, err := GetIssue(issueID)
	if err != nil {
		return nil, err
	}
	if is.ID == "" {
		return nil, fmt.Errorf("issue %s not found", issueID)
	}
	var urls []string
	for _, m := range markdownImageRe.FindAllStringSubmatch(is.Description, -1) {
		urls = append(urls, m[1])
	}
	return urls, nil
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
