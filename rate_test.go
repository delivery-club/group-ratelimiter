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
}

const (
	firstGroup  = "first_group"
	secondGroup = "second_group"
)

func TestTake(t *testing.T) {
	tests := []testCase{
		{
			firstGroupRate:  5,
			secondGroupRate: 5,
			masterRate:      100,
			tt:              t,
		},
		{
			firstGroupRate:  15,
			secondGroupRate: 5,
			masterRate:      200,
			tt:              t,
		},
		{
			firstGroupRate:  30,
			secondGroupRate: 60,
			masterRate:      100,
			tt:              t,
		},
	}

	var wg sync.WaitGroup
	for testNum, tc := range tests {
		wg.Add(1)

		go func(testNum int, tc testCase) {
			defer wg.Done()
			testRun(testNum, tc)
		}(testNum, tc)
	}
	wg.Wait()
}

func testRun(testNum int, test testCase) {
	clockMock := clock.NewMock()
	opts := []ratelimit.Option{WithoutSlack(), WithClock(clockMock)}

	rl := New(test.masterRate, opts...).
		AddGroup(firstGroup, test.firstGroupRate, opts...).
		AddGroup(secondGroup, test.secondGroupRate, opts...)

	var (
		wg    sync.WaitGroup
		count int64
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer wg.Wait()
	defer cancel()

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
			atomic.AddInt64(&count, 1)
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

	wg.Add(2)
	clockMock.AfterFunc(1*time.Second, func() {
		wg.Done()
		assert.Equal(test.tt, test.firstGroupRate, int(count), "testNum: %d, expected count: %d, actual count: %d", testNum, test.firstGroupRate, count)
	})
	clockMock.AfterFunc(2*time.Second, func() {
		wg.Done()
		assert.Equal(test.tt, 2*test.firstGroupRate, int(count), "testNum: %d, expected count: %d, actual count: %d", testNum, 2*test.firstGroupRate, count)
	})

	clockMock.Add(2 * time.Second)
}
