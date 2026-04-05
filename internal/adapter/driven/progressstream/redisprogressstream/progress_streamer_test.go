package redisprogressstream_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/st-ember/streaming-api/internal/adapter/driven/progressstream/redisprogressstream"
	"github.com/st-ember/streaming-api/internal/adapter/driven/redis"
	mockLog "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	"github.com/st-ember/streaming-api/internal/domain/progress"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedisProgressStreamer(t *testing.T) {
	t.Run("success case - multiple updates", func(t *testing.T) {
		s := miniredis.RunT(t)
		client, _ := redis.NewClient([]string{s.Addr()}, "")
		logger := mockLog.NewMockLogger(t)
		streamer := redisprogressstream.NewRedisProgressStreamer(client, logger)

		jobID := "test_job"
		ch, err := streamer.Read(t.Context(), jobID)
		require.NoError(t, err)

		prg, _ := progress.NewProgress(100)

		// Sequence of updates
		updates := []int64{25, 50, 75, 100}
		for _, val := range updates {
			prg.UpdateCurrentFrames(val)
			err = streamer.Push(t.Context(), jobID, prg)
			require.NoError(t, err)

			select {
			case received := <-ch:
				require.Equal(t, int(val), received.Percentage)
			case <-time.After(500 * time.Millisecond):
				t.Fatalf("Timed out waiting for update %d", val)
			}
		}
	})

	t.Run("context cancellation - stops goroutine and closes channel", func(t *testing.T) {
		s := miniredis.RunT(t)
		client, _ := redis.NewClient([]string{s.Addr()}, "")
		logger := mockLog.NewMockLogger(t)
		streamer := redisprogressstream.NewRedisProgressStreamer(client, logger)

		ctx, cancel := context.WithCancel(t.Context())
		jobID := "cancel_job"

		ch, err := streamer.Read(ctx, jobID)
		require.NoError(t, err)

		// Cancel context
		cancel()

		select {
		case _, ok := <-ch:
			require.False(t, ok, "Channel should be closed after context cancellation")
		case <-time.After(1 * time.Second):
			t.Fatal("Channel was not closed in time")
		}
	})

	t.Run("invalid data - logs error and continues", func(t *testing.T) {
		s := miniredis.RunT(t)
		client, _ := redis.NewClient([]string{s.Addr()}, "")
		logger := mockLog.NewMockLogger(t)
		streamer := redisprogressstream.NewRedisProgressStreamer(client, logger)

		jobID := "error_job"
		ch, err := streamer.Read(t.Context(), jobID)
		require.NoError(t, err)

		// Expect an error log for the invalid JSON
		logger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

		// Manually publish invalid JSON directly to Redis
		channelName := fmt.Sprintf("video:%s:progress", jobID)
		s.Publish(channelName, "{invalid-json}")

		// Publish a valid one immediately after
		prg, _ := progress.NewProgress(100)
		prg.UpdateCurrentFrames(50)
		streamer.Push(t.Context(), jobID, prg)

		select {
		case received := <-ch:
			require.Equal(t, 50, received.Percentage)
		case <-time.After(1 * time.Second):
			t.Fatal("Streamer stopped after receiving invalid data")
		}
	})

	t.Run("channel isolation - updates go to correct job", func(t *testing.T) {
		s := miniredis.RunT(t)
		client, _ := redis.NewClient([]string{s.Addr()}, "")
		logger := mockLog.NewMockLogger(t)
		streamer := redisprogressstream.NewRedisProgressStreamer(client, logger)

		jobA, jobB := "job_A", "job_B"
		chA, _ := streamer.Read(t.Context(), jobA)
		chB, _ := streamer.Read(t.Context(), jobB)

		prg, _ := progress.NewProgress(100)
		prg.UpdateCurrentFrames(42)

		// Push to Job A
		streamer.Push(t.Context(), jobA, prg)

		select {
		case received := <-chA:
			require.Equal(t, 42, received.Percentage)
		case <-chB:
			t.Fatal("Received update on Job B channel meant for Job A")
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Did not receive update for Job A")
		}
	})
}
