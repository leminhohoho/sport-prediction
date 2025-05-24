package scheduler

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	s := NewScheduler(42, context.Background(), false)

	if err := s.Run(
		ActionFunc(func(ctx context.Context) error {
			log.Printf("Hello ")
			return nil
		}),
		ActionFunc(func(ctx context.Context) error {
			log.Printf("World, ")
			return nil
		}),
		ActionFuncDelay{func(ctx context.Context) error {
			log.Printf("Oh wait, ")
			return nil
		}, time.Second * 1},
		ActionFuncDelay{func(ctx context.Context) error {
			log.Printf("In 3 second, some random words will be said, ")
			return nil
		}, time.Second * 1},
		Sleep(time.Second*5),
		ActionFuncsRandom{
			func(ctx context.Context) error {
				log.Printf("skibidi, ")
				return nil
			},
			func(ctx context.Context) error {
				log.Printf("toilet, ")
				return nil
			},
			func(ctx context.Context) error {
				log.Printf("sigma, ")
				return nil
			},
			func(ctx context.Context) error {
				log.Printf("Skrrt, ")
				return nil
			},
			func(ctx context.Context) error {
				log.Printf("Hail, ")
				return nil
			},
			func(ctx context.Context) error {
				log.Printf("Favela, ")
				return nil
			},
		},
		ActionFunc(func(ctx context.Context) error {
			log.Printf("I have brainrot maximum level\n")
			return nil
		}),
	); err != nil {
		t.Error(err)
	}
}
