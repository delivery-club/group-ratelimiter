package ratelimiter
//
//import (
//	"context"
//	"fmt"
//	"sync"
//	"time"
//
//	"github.com/andres-erbsen/clock"
//)
//
//func Example() {
//	const firstGroup = "firstGroup"
//	const secondGroup = "secondGroup"
//
//	rl := New(1000, WithoutSlack()).
//		AddGroup(firstGroup, 100, WithoutSlack()).
//		AddGroup(secondGroup, 200, WithoutSlack())
//
//	ch1 := make(chan string, 10)
//	ch2 := make(chan string, 10)
//
//	var wg sync.WaitGroup
//	wg.Add(2)
//	ctx := context.Background()
//
//	go func() {
//		defer wg.Done()
//		prev := time.Now()
//		for i := 0; i < 5; i++ {
//			now := rl.Take(ctx, firstGroup)
//			if i != 0 {
//				ch1 <- now.Sub(prev).String()
//			}
//			prev = now
//		}
//	}()
//
//	go func() {
//		defer wg.Done()
//		prev := time.Now()
//		for i := 0; i < 5; i++ {
//			now := rl.Take(ctx, secondGroup)
//			if i != 0 {
//				ch2 <- now.Sub(prev).String()
//			}
//			prev = now
//		}
//	}()
//	wg.Wait()
//
//	close(ch1)
//	close(ch2)
//
//	for v := range ch1 {
//		fmt.Println(v)
//	}
//
//	for v := range ch2 {
//		fmt.Println(v)
//	}
//
//	// Output:
//	// 10ms
//	// 10ms
//	// 10ms
//	// 10ms
//	// 5ms
//	// 5ms
//	// 5ms
//	// 5ms
//}
//
////nolint:govet
////goland:noinspection GoTestName
//func ExampleWithAnotherTimeWindow() {
//	const firstGroup = "firstGroup"
//
//	rl := New(1000, Per(500*time.Millisecond)).
//		AddGroup(firstGroup, 100, Per(500*time.Millisecond)) // 100 per half second or 200rps
//
//	ctx := context.Background()
//	prev := time.Now()
//	for i := 0; i < 10; i++ {
//		now := rl.Take(ctx, firstGroup)
//		if i > 0 {
//			fmt.Println(i, now.Sub(prev))
//		}
//		prev = now
//	}
//
//	// Output:
//	// 1 5ms
//	// 2 5ms
//	// 3 5ms
//	// 4 5ms
//	// 5 5ms
//	// 6 5ms
//	// 7 5ms
//	// 8 5ms
//	// 9 5ms
//}
//
//type clockDecorator struct {
//	clock.Clock
//	value string
//}
//
//func (c clockDecorator) Sleep(t time.Duration) {
//	fmt.Println("decorated sleep", c.value)
//	time.Sleep(t)
//}
//
////nolint:govet
////goland:noinspection GoTestName
//func ExampleWithClockDecorator() {
//	const firstGroup = "firstGroup"
//
//	cl := clockDecorator{Clock: clock.New(), value: "master"}
//	cl2 := clockDecorator{Clock: clock.New(), value: "group"}
//
//	rl := New(1000, WithClock(cl)).
//		AddGroup(firstGroup, 100, WithClock(cl2))
//
//	ctx := context.Background()
//	for i := 0; i < 2; i++ {
//		rl.Take(ctx, firstGroup)
//	}
//
//	// Output:
//	// decorated sleep master
//	// decorated sleep group
//	// decorated sleep master
//	// decorated sleep group
//}
