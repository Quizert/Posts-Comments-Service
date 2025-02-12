package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Quizert/PostCommentService/internal/models"
	"github.com/Quizert/PostCommentService/internal/service"
)

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	s := service.NewSubscriptionService()
	ctx := context.Background()

	ch, err := s.CreateSubscription(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, ch)

	comment := &models.Comment{ID: 10, PostID: 1, Payload: "Hello"}
	go func() {
		_ = s.Notify(ctx, comment)
	}()

	c := <-ch
	assert.Equal(t, comment, c)
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
	s := service.NewSubscriptionService()
	ctx := context.Background()

	fakeChan := make(chan *models.Comment)
	err := s.DeleteSubscription(ctx, 999, fakeChan)
	require.NoError(t, err, "deleting non-existent subscription should not return error")

	ch, err := s.CreateSubscription(ctx, 1)
	require.NoError(t, err)

	err = s.DeleteSubscription(ctx, 1, ch)
	require.NoError(t, err)

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after DeleteSubscription")

	err = s.DeleteSubscription(ctx, 1, ch)
	require.NoError(t, err, "second delete should do nothing and not fail")
}

func TestSubscriptionService_Notify(t *testing.T) {
	s := service.NewSubscriptionService()
	ctx := context.Background()

	err := s.Notify(ctx, &models.Comment{PostID: 999, Payload: "Nothing"})
	require.NoError(t, err)

	ch1, err := s.CreateSubscription(ctx, 1)
	require.NoError(t, err)
	ch2, err := s.CreateSubscription(ctx, 1)
	require.NoError(t, err)

	comment := &models.Comment{ID: 10, PostID: 1, Payload: "Broadcast"}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c := <-ch1
		assert.Equal(t, comment, c, "ch1 should get the comment")
	}()
	go func() {
		defer wg.Done()
		c := <-ch2
		assert.Equal(t, comment, c, "ch2 should get the comment")
	}()

	go func() {
		err := s.Notify(ctx, comment)
		assert.NoError(t, err)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// всё ок
	case <-time.After(time.Second):
		t.Fatal("timeout: channels did not receive the message in time")
	}
}
