package handler

type UploadVideoResponse struct {
	VideoID    string `json:"video_id"`
	JobID      string `json:"job_id"`
	Status     string `json:"status"`
	ResourceID string `json:"resource_id"`
}
