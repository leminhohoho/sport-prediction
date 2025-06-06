package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/leminhohoho/sport-prediction/runner/helpers"
)

type Action interface {
	Do(context.Context, *Scheduler)
}

type Scheduler struct {
	ctx      context.Context
	parallel int
	repeat   bool

	wg        sync.WaitGroup
	semaphore chan struct{}
	errChan   chan error
}

// Create a new scheduler, parallel determine the maximum amount of processes that
// can be run simultaneously, values below 2 mean no parallelism. Repeat will
// determine if the scheduler run on repeat or not.
func NewScheduler(parallel int, ctx context.Context, repeat bool) *Scheduler {
	return &Scheduler{
		parallel: parallel,
		ctx:      ctx,
		repeat:   repeat,

		semaphore: make(chan struct{}, parallel),
		errChan:   make(chan error, parallel),
	}
}

// Execute a list of actions sequentially
func (s *Scheduler) Run(actions ...Action) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case err := <-s.errChan:
		return err
	default:
		for _, action := range actions {
			action.Do(s.ctx, s)
		}

		s.wg.Wait()

		if s.repeat {
			return s.Run(actions...)
		}

		return nil
	}
}

type Async struct {
	a Action
}

func (as Async) Do(ctx context.Context, s *Scheduler) {
	go as.a.Do(ctx, s)
}

// Execute a normal function as an action
type ActionFunc func(context.Context) error

func (f ActionFunc) Do(ctx context.Context, s *Scheduler) {
	s.semaphore <- struct{}{}
	s.wg.Add(1)
	defer func() {
		<-s.semaphore
		s.wg.Done()
	}()

	if err := f(ctx); err != nil {
		s.errChan <- err
	}
}

// Execute a normal but with a delay that fluctuate between 0 and the specified duration
type ActionFuncDelay struct {
	funcToRun func(context.Context) error
	duration  time.Duration
}

func (f ActionFuncDelay) Do(ctx context.Context, s *Scheduler) {
	s.semaphore <- struct{}{}
	s.wg.Add(1)
	defer func() {
		<-s.semaphore
		s.wg.Done()
	}()

	time.Sleep(helpers.GetRandomTime(time.Duration(0), f.duration))
	if err := f.funcToRun(ctx); err != nil {
		s.errChan <- err
	}
}

// Take a list of normal function and run it in a randomized order, the randomness is
// controlled by the randomness variable
type ActionFuncsRandom []func(context.Context) error

func (fs ActionFuncsRandom) Do(ctx context.Context, s *Scheduler) {
	s.semaphore <- struct{}{}
	s.wg.Add(1)
	defer func() {
		<-s.semaphore
		s.wg.Done()
	}()

	randomizedOrder := helpers.RandomizeCyclicGroup(len(fs))
	var err error

	for _, i := range randomizedOrder {
		if err = fs[i](ctx); err != nil {
			s.errChan <- err
		}
	}
}

// Pause for a time duration
type Sleep time.Duration

func (timer Sleep) Do(ctx context.Context, s *Scheduler) {
	s.semaphore <- struct{}{}
	s.wg.Add(1)
	defer func() {
		<-s.semaphore
		s.wg.Done()
	}()

	time.Sleep(time.Duration(timer))
}
