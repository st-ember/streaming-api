package handler

import (
	"encoding/json"
	"net/http"

	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit request at 1 GB
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*1024)

	// Parse form request
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Errorf("parse form request: %v", err)
		http.Error(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		h.logger.Errorf("find video file from form request: %v", err)
		http.Error(w, "missing video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	description := r.FormValue("description")

	// Assemble usecase input
	input := videoapp.UploadVideoInput{
		Title:        title,
		Description:  description,
		FileName:     header.Filename,
		VideoContent: file,
	}

	// Execute usecase
	result, err := h.uploadVideoUC.Execute(r.Context(), input)
	if err != nil {
		h.logger.Errorf("execute upload video usecase: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Assemble response
	response := UploadVideoResponse{
		VideoID:    result.Video.ID,
		JobID:      result.Job.ID,
		Status:     string(result.Video.Status),
		ResourceID: result.Video.ResourceID,
	}

	// Set header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("encode upload video response: %v", err)
		return
	}
}
