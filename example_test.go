package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
)

type config struct {
	Master    int
	GroupRate map[string]int
}

func (c config) GroupRates() map[string]int {
	return c.GroupRate
}

func (c config) MasterRate() int {
	return c.Master
}

func Example() {
	const firstGroup = "firstGroup"
	const secondGroup = "second_group"

	rl := NewRateLimiterGroup(config{
		Master: 1000,
		GroupRate: map[string]int{
			firstGroup:  100,
			secondGroup: 200,
		},
	}, WithoutSlack())

	ch1 := make(chan string, 10)
	ch2 := make(chan string, 10)

	var wg sync.WaitGroup
	wg.Add(2)
	ctx := context.Background()

	go func() {
		defer wg.Done()
		prev := time.Now()
		for i := 0; i < 5; i++ {
			now := rl.Take(ctx, firstGroup)
			if i != 0 {
				ch1 <- now.Sub(prev).String()
			}
			prev = now
		}
	}()

	go func() {
		defer wg.Done()
		prev := time.Now()
		for i := 0; i < 5; i++ {
			now := rl.Take(ctx, secondGroup)
			if i != 0 {
				ch2 <- now.Sub(prev).String()
			}
			prev = now
		}
	}()
	wg.Wait()

	close(ch1)
	close(ch2)

	for v := range ch1 {
		fmt.Println(v)
	}

	for v := range ch2 {
		fmt.Println(v)
	}

	// Output:
	// 10ms
	// 10ms
	// 10ms
	// 10ms
	// 5ms
	// 5ms
	// 5ms
	// 5ms
}

//nolint:govet
func ExampleWithAnotherTimeWindow() {
	const firstGroup = "firstGroup"

	rl := NewRateLimiterGroup(config{
		Master:    1000,
		GroupRate: map[string]int{firstGroup: 100}, // 100 per half second, 200 Hz
	}, Per(500*time.Millisecond))

	ctx := context.Background()
	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take(ctx, firstGroup)
		if i > 0 {
			fmt.Println(i, now.Sub(prev))
		}
		prev = now
	}

	// Output:
	// 1 5ms
	// 2 5ms
	// 3 5ms
	// 4 5ms
	// 5 5ms
	// 6 5ms
	// 7 5ms
	// 8 5ms
	// 9 5ms
}

type clockDecorator struct {
	clock.Clock
}

func (c clockDecorator) Sleep(t time.Duration) {
	fmt.Println("decorated sleep")
	time.Sleep(t)
}

//nolint:govet
func ExampleWithClockDecorator() {
	const firstGroup = "firstGroup"

	cl := clockDecorator{clock.New()}

	rl := NewRateLimiterGroup(config{
		Master:    1000,
		GroupRate: map[string]int{firstGroup: 100},
	}, WithClock(cl))

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		rl.Take(ctx, firstGroup)
	}

	// Output:
	// decorated sleep
	// decorated sleep
	// decorated sleep
	// decorated sleep
}
