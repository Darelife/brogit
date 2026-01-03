package api

import "time"

// ChangeType defines the type of change (File modification, Command, etc.)
type ChangeType string

const (
	ChangeTypeFileEdit ChangeType = "file_edit"
	ChangeTypeCommit   ChangeType = "commit"
)

// DiffEntry represents a single change or command from a client
type DiffEntry struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Timestamp time.Time  `json:"timestamp"`
	Type      ChangeType `json:"type"`
	FilePath  string     `json:"file_path,omitempty"`
	Content   string     `json:"content,omitempty"` // Could be a diff or full content
	Params    string     `json:"params,omitempty"`  // Extra params for commands
}

// PushRequest is the body for the /push endpoint
type PushRequest struct {
	Entry DiffEntry `json:"entry"`
}

// PushResponse is the response for the /push endpoint
type PushResponse struct {
	Status string `json:"status"`
}
