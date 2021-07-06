package ratelimiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/ratelimit"
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

	wg.Wait()
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

	rl := NewRateLimiter(limitConf, ratelimit.WithoutSlack)

	var count int64
	for i := 0; i < int(test.testCount); i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			rl.Take(firstGroup)
			atomic.AddInt64(&count, 1)
		}()

		go func() {
			defer wg.Done()
			rl.Take(secondGroup)
		}()
	}

	wg.Add(2)
	time.AfterFunc(time.Second, func() {
		defer wg.Done()
		if test.groupRate != int(count) {
			test.tt.Errorf("expected count: %d, actual count: %d", test.groupRate, count)
		}
	})

	time.AfterFunc(2*time.Second, func() {
		defer wg.Done()
		if test.groupRate*2 != int(count) {
			test.tt.Errorf("expected count: %d, actual count: %d", 2*test.groupRate, count)
		}
	})

	wg.Wait()
}
