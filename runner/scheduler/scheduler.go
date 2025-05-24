package scheduler

import (
	"context"
	"time"

	"github.com/leminhohoho/sport-prediction/helpers"
)

type Action interface {
	Do(context.Context, Scheduler) error
}

type Scheduler struct {
	randomness int
	ctx        context.Context
	repeat     bool
}

// Create a new scheduler, randomness will determine how unpredictable the scheduler will behave,
// and repeat will determine if the scheduler run on repeat or not.
func NewScheduler(randomness int, ctx context.Context, repeat bool) *Scheduler {
	return &Scheduler{
		randomness: randomness,
		ctx:        ctx,
		repeat:     repeat,
	}
}

// Execute a list of actions sequentially
func (s *Scheduler) Run(actions ...Action) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		var err error
		for _, action := range actions {
			if err = action.Do(s.ctx, *s); err != nil {
				return err
			}
		}

		if s.repeat {
			return s.Run(actions...)
		}

		return nil
	}
}

// Execute a normal function as an action
type ActionFunc func(context.Context) error

func (f ActionFunc) Do(ctx context.Context, s Scheduler) error {
	return f(ctx)
}

// Execute a normal but with a delay that fluctuate between 0 and the specified duration
type ActionFuncDelay struct {
	funcToRun func(context.Context) error
	duration  time.Duration
}

func (f ActionFuncDelay) Do(ctx context.Context, s Scheduler) error {
	time.Sleep(helpers.GetRandomTime(time.Duration(0), f.duration))
	return f.funcToRun(ctx)
}

// Take a list of normal function and run it in a randomized order, the randomness is
// controlled by the randomness variable
type ActionFuncsRandom []func(context.Context) error

func (fs ActionFuncsRandom) Do(ctx context.Context, s Scheduler) error {
	randomizedOrder := helpers.RandomizeCyclicGroup(len(fs), s.randomness)
	var err error

	for _, i := range randomizedOrder {
		if err = fs[i](ctx); err != nil {
			return err
		}
	}

	return nil
}

// Pause for a time duration
type Sleep time.Duration

func (timer Sleep) Do(ctx context.Context, s Scheduler) error {
	time.Sleep(time.Duration(timer))
	return nil
}
