package redisprogressstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/st-ember/streaming-api/internal/adapter/driven/redis"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/progressstream"
	"github.com/st-ember/streaming-api/internal/domain/progress"
)

type RedisProgressStreamer struct {
	Client *redis.Client
	logger log.Logger
}

// NewRedisProgressStreamer initializes the RedisProgressStreamer struct
func NewRedisProgressStreamer(client *redis.Client, logger log.Logger) progressstream.ProgressStreamer {
	return &RedisProgressStreamer{Client: client, logger: logger}
}

// Push publishes progress objects to redis for clients to consume in real-time
func (p *RedisProgressStreamer) Push(ctx context.Context, jobID string, prg *progress.Progress) error {
	// Marshal progress object
	data, err := json.Marshal(prg)
	if err != nil {
		return fmt.Errorf("marshal progress for job %s: %w", jobID, err)
	}

	// Push to redis
	if err := p.Client.Rdb.Publish(ctx, p.buildChannel(jobID), data).Err(); err != nil {
		return fmt.Errorf("publish progress to redis for job %s: %w", jobID, err)
	}

	return nil
}

// Read returns a channel with continuously updated progress objects for a client connection to consume
func (p *RedisProgressStreamer) Read(ctx context.Context, jobID string) (<-chan *progress.Progress, error) {
	ch := make(chan *progress.Progress)
	redisCh := p.buildChannel(jobID)
	pubsub := p.Client.Rdb.Subscribe(ctx, redisCh)

	// Waits for error or first confirmation message
	if _, err := pubsub.Receive(ctx); err != nil {
		pubsub.Close()
		return nil, fmt.Errorf("subscribe to redis channel %s: %w", redisCh, err)
	}

	go func() {
		defer close(ch)
		defer pubsub.Close()

		redisMsgCh := pubsub.Channel()

		// Push all received progress to go channel
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-redisMsgCh:
				if !ok { // redis channel
					return
				}

				var prg progress.Progress
				if err := json.Unmarshal([]byte(msg.Payload), &prg); err != nil {
					p.logger.Errorf(ctx, "unmarshal progress for job %s: %v", jobID, err)
					continue
				}

				select {
				case ch <- &prg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// buildChannel builds channel name for methods in RedisProgressStreamer
func (p *RedisProgressStreamer) buildChannel(jobID string) string {
	return fmt.Sprintf("video:%s:progress", jobID)
}
