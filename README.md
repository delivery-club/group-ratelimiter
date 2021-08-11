# Group Ratelimiter

Данный пакет предназначен для ограничения доступа к ресурсу между несколькими группами потребителей.

Под собой библиотека использует https://github.com/uber-go/ratelimit

```go
import (
	"fmt"
	"sync"
	"time"
)

func main() {
	const firstGroup = "firstGroup"
	const secondGroup = "secondGroup"

	rl := New(1000). // global limit 1000 rps
	    AddGroup(firstGroup, 100). // limit of the first group 100 rps
	    AddGroup(secondGroup, 200) // limit of the second group 200 rps

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

    // where first four values of the output are time waited before requests for firstGroup
    // and second four values time waited before requests for secondGroup

	// ms - millisecond or one thousandth of a second
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
