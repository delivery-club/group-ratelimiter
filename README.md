# Group Ratelimiter

Данный пакет предназначен для ограничения доступа к ресурсу между несколькими группами потребителей.

Под собой библиотека использует https://github.com/uber-go/ratelimit

```go
import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/ratelimit"
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

func main() {
	const firstGroup = "firstGroup"
	const secondGroup = "secondGroup"

	rl := NewRateLimiterGroup(config{
		Master: 1000, // global limit 1000 rps
		GroupRate: map[string]int{
			firstGroup:  100, // limit of the first group 100 rps
			secondGroup: 200, // limit of the second group 200 rps
		},
	}, ratelimit.WithoutSlack)

	ch1 := make(chan string, 10)
	ch2 := make(chan string, 10)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		prev := time.Now()
		for i := 0; i < 5; i++ {
			now := rl.Take(firstGroup)
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
			now := rl.Take(secondGroup)
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
		fmt.Println(v) // time waited before request for firstGroup
	}

	for v := range ch2 {
		fmt.Println(v) // time waited before request for secondGroup
	}

	// where ms - millisecond or one thousandth of a second
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
```
