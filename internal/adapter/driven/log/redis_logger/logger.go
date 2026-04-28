package redislogger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	logPort "github.com/st-ember/streaming-api/internal/application/ports/log"

	"github.com/st-ember/streaming-api/internal/adapter/driven/redis"
)

type RedisLogger struct {
	Client *redis.Client
}

func NewRedisLogger(client *redis.Client) logPort.Logger {
	return &RedisLogger{Client: client}
}

const (
	errLogChannel  = "logs:streaming-api:error"
	warnLogChannel = "logs:streaming-api:warn"
	infoLogChannel = "logs:streaming-api:info"
)

func (l *RedisLogger) Errorf(ctx context.Context, category logPort.LogCategory, sourceID string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	lMsg := LogMessage{
		Category: category.String(),
		Level:    LevelError.String(),
		Message:  msg,
		SourceID: sourceID,
	}

	b, err := json.Marshal(lMsg)
	if err != nil {
		log.Printf("ERROR: failed to marshal log message")
	}

	l.Client.Rdb.Publish(ctx, errLogChannel, b)
}

func (l *RedisLogger) Warnf(ctx context.Context, category logPort.LogCategory, sourceID string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	lMsg := LogMessage{
		Category: category.String(),
		Level:    LevelWarn.String(),
		Message:  msg,
		SourceID: sourceID,
	}

	b, err := json.Marshal(lMsg)
	if err != nil {
		log.Printf("ERROR: failed to marshal log message")
	}

	l.Client.Rdb.Publish(ctx, warnLogChannel, b)
}

func (l *RedisLogger) Infof(ctx context.Context, category logPort.LogCategory, sourceID string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	lMsg := LogMessage{
		Category: category.String(),
		Level:    LevelInfo.String(),
		Message:  msg,
		SourceID: sourceID,
	}

	b, err := json.Marshal(lMsg)
	if err != nil {
		log.Printf("ERROR: failed to marshal log message")
	}

	l.Client.Rdb.Publish(ctx, infoLogChannel, b)
}
