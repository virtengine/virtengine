package jira

// PlainText returns a best-effort plain text representation of the comment body.
func (c *Comment) PlainText() string {
	if c == nil || c.Body == nil {
		return ""
	}
	return c.Body.PlainText()
}

// IsInternal returns true when the comment is marked as internal.
func (c *Comment) IsInternal() bool {
	if c == nil {
		return false
	}
	for _, prop := range c.Properties {
		if prop.Key != "sd.public.comment" {
			continue
		}
		switch value := prop.Value.(type) {
		case map[string]interface{}:
			if internal, ok := value["internal"].(bool); ok {
				return internal
			}
		case map[string]bool:
			if internal, ok := value["internal"]; ok {
				return internal
			}
		}
	}
	return false
}

// NewCommentRequest builds a Jira comment request with optional internal flag.
func NewCommentRequest(body string, internal bool) *AddCommentRequest {
	req := &AddCommentRequest{
		Body: NewADFDocument(body),
	}
	if internal {
		req.Properties = []CommentProperty{{
			Key:   "sd.public.comment",
			Value: map[string]bool{"internal": true},
		}}
	}
	return req
}
