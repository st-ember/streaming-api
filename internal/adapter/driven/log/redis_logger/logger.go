package redislogger

import (
	"context"
	"fmt"

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

func (l *RedisLogger) Errorf(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.Client.Rdb.Publish(ctx, errLogChannel, msg)
}

func (l *RedisLogger) Warnf(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.Client.Rdb.Publish(ctx, warnLogChannel, msg)
}

func (l *RedisLogger) Infof(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.Client.Rdb.Publish(ctx, infoLogChannel, msg)
}
