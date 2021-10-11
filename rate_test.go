package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
)

type testCase struct {
	firstGroupRate  int
	secondGroupRate int
	masterRate      int
	tt              *testing.T
	wg              *sync.WaitGroup
	clock           *clock.Mock
	count           int32
	name            string
}

const (
	firstGroup  = "first_group"
	secondGroup = "second_group"
)

func TestCase1(t *testing.T) {
	RunTestCase(testCase{
		firstGroupRate:  20,
		secondGroupRate: 60,
		masterRate:      100,
		tt:              t,
		name:            "case_1",
	}, t)
}

func TestCase2(t *testing.T) {
	RunTestCase(testCase{
		firstGroupRate:  60,
		secondGroupRate: 20,
		masterRate:      100,
		tt:              t,
		name:            "case_2",
	}, t)
}

func RunTestCase(tc testCase, t *testing.T) {
	t.Run(tc.name, func(t *testing.T) {
		tc.clock = clock.NewMock()
		tc.wg = &sync.WaitGroup{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		defer tc.wg.Wait()

		tc.testRun(ctx)

		tc.clock.Add(3 * time.Second)
	})
}

func (t *testCase) testRun(ctx context.Context) {
	opts := []ratelimit.Option{WithoutSlack(), WithClock(t.clock)}

	rl := New(t.masterRate, opts...).
		AddGroup(firstGroup, t.firstGroupRate, opts...).
		AddGroup(secondGroup, t.secondGroupRate, opts...)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		wg.Done()
		for {
			rl.Take(ctx, firstGroup)
			select {
			case <-ctx.Done():
				return
			default:
			}
			atomic.AddInt32(&t.count, 1)
		}
	}()
	go func() {
		wg.Done()
		for {
			rl.Take(ctx, secondGroup)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
	wg.Wait()

	t.after(1*time.Second, 1*t.firstGroupRate)
	t.after(2*time.Second, 2*t.firstGroupRate)
	t.after(3*time.Second, 3*t.firstGroupRate)
}

func (t *testCase) after(d time.Duration, count int) {
	t.wg.Add(1)
	t.clock.AfterFunc(d, func() {
		assert.Equal(t.tt, count, int(t.count), "test: %s, expected count: %d, actual count: %d", t.name, count, int(t.count))
		t.wg.Done()
	})
}

func TestContext(t *testing.T) {
	t.Run("context_cancel_before_take", func(t *testing.T) {
		rl := New(1000, WithSlack(0)).
			AddGroup(firstGroup, 1, Per(time.Hour))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		now := time.Now()
		for i := 0; i < 10; i++ {
			if d := rl.Take(ctx, firstGroup).Sub(now); d.Nanoseconds() > time.Microsecond.Nanoseconds() {
				t.Fatalf("not equal duration='%s'", d.String())
			}
			now = time.Now()
		}

		//take undefined group
		_ = rl.Take(context.Background(), secondGroup)
	})

	t.Run("context_cancel_on_group", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cl := clockMock{counter: ptfOfInt(0), Clock: clock.New(), cancelFunc: cancel}
		rl := New(100, WithoutSlack(), WithClock(cl), Per(time.Second)).
			AddGroup(firstGroup, 10, Per(time.Second))

		rl.Take(ctx, firstGroup)

		now := time.Now()
		if d := rl.Take(ctx, firstGroup).Sub(now); time.Second/100 != d.Round(5*time.Microsecond) {
			t.Fatalf("must be equal to master rate='%s'", d)
		}
	})
}

type clockMock struct {
	clock.Clock
	counter    *int
	cancelFunc context.CancelFunc
}

func (c clockMock) Sleep(duration time.Duration) {
	c.Clock.Sleep(duration)

	*c.counter++
	if *c.counter > 1 {
		c.cancelFunc()
	}
}

func ptfOfInt(i int) *int { return &i }
