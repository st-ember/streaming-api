package redislogger

type LogMessage struct {
	Category string `json:"category"`
	Level    string `json:"level"`
	Message  string `json:"message"`
	SourceID string `json:"source_id,omitempty"`
}

type LogLevel string

const (
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

func (ll LogLevel) String() string {
	return string(ll)
}
