package log

type LogCategory string

const (
	CategoryDefault LogCategory = "default"
	CategoryVideo   LogCategory = "video"
	CategoryJob     LogCategory = "job"
)

func (lc LogCategory) String() string {
	return string(lc)
}
