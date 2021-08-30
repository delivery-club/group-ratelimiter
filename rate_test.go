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
