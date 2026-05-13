package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
	"github.com/redis/go-redis/v9"
)

type QuizTimerService struct {
	queries    *db.Queries
	session    *QuizSessionService
	cache      *CacheService
	redis      *redis.Client
	done       chan struct{}
	subscribed chan struct{}
}

func NewQuizTimerService(queries *db.Queries, session *QuizSessionService, cache *CacheService, redisClient *redis.Client) *QuizTimerService {
	return &QuizTimerService{
		queries:    queries,
		session:    session,
		cache:      cache,
		redis:      redisClient,
		done:       make(chan struct{}),
		subscribed: make(chan struct{}),
	}
}

func (s *QuizTimerService) StartTimerSubscription(ctx context.Context) error {
	ps := s.redis.PSubscribe(ctx, "__keyevent@0__:expired")
	defer ps.Close()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("quiz timer: recovered from panic: %v", r)
			}
		}()

		ch := ps.Channel()
		for {
			select {
			case <-ctx.Done():
				close(s.subscribed)
				return
			case <-s.done:
				close(s.subscribed)
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg == nil {
					continue
				}
				key := msg.Payload
				if after, ok := strings.CutPrefix(key, "quiz:timer:"); ok {
					attemptIDStr := after
					attemptID, err := uuid.Parse(attemptIDStr)
					if err != nil {
						log.Printf("quiz timer: failed to parse attempt ID %s: %v", attemptIDStr, err)
						continue
					}
					s.expireAttempt(ctx, attemptID)
				}
			}
		}
	}()

	return nil
}

func (s *QuizTimerService) expireAttempt(ctx context.Context, attemptID uuid.UUID) {
	attempt, err := s.queries.GetAttemptByID(ctx, attemptID)
	if err != nil {
		log.Printf("quiz timer: failed to get attempt %s: %v", attemptID, err)
		return
	}

	if attempt.CompletedAt.Valid {
		log.Printf("quiz timer: attempt %s already completed", attemptID)
		return
	}

	log.Printf("quiz timer: auto-completing attempt %s for user %s", attemptID, attempt.UserID)

	_, err = s.session.CompleteSessionByAttemptID(ctx, attemptID)
	if err != nil {
		log.Printf("quiz timer: failed to complete attempt %s: %v", attemptID, err)
		return
	}
}

func (s *QuizTimerService) StartCronJob(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.expireStaleAttempts(ctx)
		case <-ctx.Done():
			return
		case <-s.done:
			return
		}
	}
}

func (s *QuizTimerService) expireStaleAttempts(ctx context.Context) {
	attempts, err := s.queries.GetStaleAttempts(ctx)
	if err != nil {
		log.Printf("quiz timer: failed to get stale attempts: %v", err)
		return
	}
	log.Printf("quiz timer: found %d stale attempts", len(attempts))
	for _, attempt := range attempts {
		s.expireAttempt(ctx, attempt.ID)
	}
}

func (s *QuizTimerService) Shutdown() {
	close(s.done)
}
