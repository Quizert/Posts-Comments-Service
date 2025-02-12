package service

import (
	"context"
	"github.com/Quizert/PostCommentService/internal/models"
	"sync"
)

type SubscriptionService struct {
	commentChannels map[int][]chan *models.Comment
	mu              sync.Mutex
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{commentChannels: map[int][]chan *models.Comment{}}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, postID int) (chan *models.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *models.Comment)
	s.commentChannels[postID] = append(s.commentChannels[postID], ch)

	return ch, nil
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, postID int, ch chan *models.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chans, ok := s.commentChannels[postID]
	if !ok {
		return nil
	}
	for i, c := range chans {
		if c == ch {
			close(ch)
			n := len(chans) - 1
			s.commentChannels[postID][i] = chans[n]
			s.commentChannels[postID] = s.commentChannels[postID][:n]

			if len(s.commentChannels[postID]) == 0 {
				delete(s.commentChannels, postID)
			}

			return nil
		}
	}
	return nil
}

func (s *SubscriptionService) Notify(ctx context.Context, comment *models.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chans, ok := s.commentChannels[comment.PostID]
	if !ok {
		return nil
	}

	for _, ch := range chans {
		ch <- comment
	}

	return nil
}
