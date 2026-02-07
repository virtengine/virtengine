// Package jira provides Jira Service Desk integration for VirtEngine.
//
// This file implements Atlassian Document Format (ADF) helpers for Jira API v3.
package jira

import (
	"encoding/json"
	"strings"
)

// ADFDocument represents an Atlassian Document Format document.
type ADFDocument struct {
	Type    string    `json:"type"`
	Version int       `json:"version"`
	Content []ADFNode `json:"content,omitempty"`
}

// ADFNode represents a node in an ADF document.
type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []ADFNode `json:"content,omitempty"`
}

// NewADFDocument builds a basic ADF document from plain text.
func NewADFDocument(text string) *ADFDocument {
	doc := &ADFDocument{
		Type:    "doc",
		Version: 1,
	}
	if text == "" {
		return doc
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		para := ADFNode{Type: "paragraph"}
		if line != "" {
			para.Content = []ADFNode{{Type: "text", Text: line}}
		}
		doc.Content = append(doc.Content, para)
	}

	return doc
}

// PlainText returns a best-effort plain text representation of the document.
func (d *ADFDocument) PlainText() string {
	if d == nil {
		return ""
	}
	if len(d.Content) == 0 {
		return ""
	}

	lines := make([]string, 0, len(d.Content))
	for _, node := range d.Content {
		lines = append(lines, nodePlainText(node))
	}
	return strings.Join(lines, "\n")
}

func nodePlainText(node ADFNode) string {
	if node.Type == "text" {
		return node.Text
	}
	if node.Type == "hardBreak" {
		return "\n"
	}
	if len(node.Content) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, child := range node.Content {
		sb.WriteString(nodePlainText(child))
	}
	return sb.String()
}

// UnmarshalJSON supports both Jira ADF objects and legacy string bodies.
func (d *ADFDocument) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == 34 {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		*d = *NewADFDocument(text)
		return nil
	}

	var alias struct {
		Type    string    `json:"type"`
		Version int       `json:"version"`
		Content []ADFNode `json:"content,omitempty"`
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*d = ADFDocument(alias)
	if d.Type == "" {
		d.Type = "doc"
	}
	if d.Version == 0 {
		d.Version = 1
	}
	return nil
}
