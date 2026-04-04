package progress

type ProgressStatus string

const (
	StatusContinue ProgressStatus = "continue"
	StatusEnd      ProgressStatus = "end"
	StatusError    ProgressStatus = "error"
)
