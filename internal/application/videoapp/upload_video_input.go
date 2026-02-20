package videoapp

import "io"

type UploadVideoInput struct {
	Title        string
	Description  string
	FileName     string
	VideoContent io.Reader
}
