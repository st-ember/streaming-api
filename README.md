# Streaming API

A Go-based API for transcoding and streaming video content.

This service provides a RESTful interface to manage video resources. Its core responsibilities are:

1.  **Transcoding:** Ingesting video files and converting them into DASH compliant manifests for adaptive bitrate streaming.
2.  **Streaming:** Serving the transcoded video segments over HTTP.

## Getting Started

### Prerequisites
- Go (version 1.18 or higher recommended)
- (Optional) FFMPEG, if transcoding logic requires it as an external dependency.

## Project Structure

This project adheres to a Hexagonal Architecture, organizing code into distinct layers based on their responsibilities:

*   `cmd/`: Contains the entry points for executable applications. `cmd/server/main.go` is the main API server.
*   `internal/`: Houses the core application logic and components not intended for external consumption.
    *   `internal/domain/`: The heart of the application, containing core business entities (`Video`, `Job`) and business rules. This layer is pure and has no external dependencies.
    *   `internal/application/`: Orchestrates domain entities to perform use cases (e.g., creating a video). It defines the ports (interfaces) for external concerns like databases.
    *   `internal/adapters/`: Provides implementations (adapters) for the ports defined in the application layer. This is where database logic (`repository`) and connections to external services (`storage`) reside.

## API Endpoints

The following table outlines the available API endpoints.

| Method | Path                  | Description                                              |
|--------|-----------------------|----------------------------------------------------------|
| `POST` | `/api/video`         | Creates a new video resource and starts a transcode job. |
| `GET`  | `/api/video/{page}`  | Lists all available video resources, with pagination.    |
| `GET`  | `/api/video/{videoId}`| Retrieves details and status for a specific video.       |
| `PUT`  | `/api/video/{videoId}`| Updates a video's metadata (e.g., title).                |
| `DELETE`| `/api/video/{videoId}`| Deletes a video manifest and all associated files.      |
| `GET`  | `/api/stream/{videoId}/manifest.mpd` | Retrieves the DASH manifest for a video.  |
