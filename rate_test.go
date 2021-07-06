package ratelimiter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andres-erbsen/clock"
)

type testCase struct {
	groupRate       int
	secondGroupRate int
	masterRate      int64
	testCount       int64
	tt              *testing.T
}

type testConf struct {
	masterRate      int
	groupRate       int
	secondGroupRate int
}

const (
	firstGroup  = "first_group"
	secondGroup = "second_group"
)

func (t *testConf) MasterRate() int {
	return t.masterRate
}

func (t *testConf) GroupRates() map[string]int {
	return map[string]int{
		firstGroup:  t.groupRate,
		secondGroup: t.secondGroupRate,
	}
}

func TestTake(t *testing.T) {
	tests := []testCase{
		{
			groupRate:       5,
			secondGroupRate: 5,
			masterRate:      100,
			testCount:       100,
			tt:              t,
		},
		{
			groupRate:       15,
			secondGroupRate: 5,
			masterRate:      200,
			testCount:       100,
			tt:              t,
		},
		{
			groupRate:       30,
			secondGroupRate: 60,
			masterRate:      100,
			testCount:       100,
			tt:              t,
		},
	}

	var wg sync.WaitGroup
	for _, tc := range tests {
		wg.Add(1)

		go func(tc testCase) {
			defer wg.Done()
			testRun(tc)
		}(tc)
	}
}

func testRun(test testCase) {
	var (
		wg        sync.WaitGroup
		limitConf = &testConf{
			masterRate:      int(test.masterRate),
			groupRate:       test.groupRate,
			secondGroupRate: test.secondGroupRate,
		}
	)

	clockMock := clock.NewMock()
	rl := NewRateLimiterGroup(limitConf, WithoutSlack(), WithClock(clockMock))
	ctx := context.Background()

	var count int64
	for i := 0; i < int(test.testCount); i++ {
		wg.Add(2)
		go func() {
			wg.Done()
			rl.Take(ctx, firstGroup)
			atomic.AddInt64(&count, 1)
		}()

		go func() {
			wg.Done()
			rl.Take(ctx, secondGroup)
		}()
	}

	wg.Add(1)
	clockMock.AfterFunc(1*time.Second, func() {
		defer wg.Done()
		if test.groupRate != int(count) {
			test.tt.Errorf("expected count: %d, actual count: %d", test.groupRate, count)
		}
	})

	wg.Add(1)
	clockMock.AfterFunc(2*time.Second, func() {
		defer wg.Done()
		if 2*test.groupRate != int(count) {
			test.tt.Errorf("expected count: %d, actual count: %d", 2*test.groupRate, count)
		}
	})

	clockMock.Add(2 * time.Second)
	wg.Wait()
}
