package handler

import "time"

type GetVideoInfoResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	SourceFilename string    `json:"source_filename"`
	ResourceID     string    `json:"resource_id"`
	Status         string    `json:"status"`
	Duration       float64   `json:"duration_seconds"`
	ManifestPath   string    `json:"manifest_path,omitempty"`
	ErrorMsg       string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
