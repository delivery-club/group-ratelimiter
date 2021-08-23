package ratelimiter

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"go.uber.org/ratelimit"
)

// Приоритизация групп возможна засчет разницы между их лимитами,
// чем выше лимит группы, тем выше ее приоритет,
// тем не менее общее количество запросов будет ограниченно основным лимитом (MasterLimit)
type GroupLimiter interface {
	Take(context context.Context, groupName string) time.Time
	AddGroup(groupName string, rate int, opts ...ratelimit.Option) GroupLimiter
}

// groupLimiter описание условий:
// 1. есть лимиты группы
// 2. есть общий лимит
// 3. группа должна уметь использовать часть лимитов общих, с возможностью приоритезации
// 4. TODO рассмотреть возможность автонаращивание лимитов группами при неиспользовании их одной из групп
type groupLimiter struct {
	masterLimit   ratelimit.Limiter
	groupLimiters map[string]ratelimit.Limiter
}

func (gr *groupLimiter) Take(ctx context.Context, group string) time.Time {
	if gl, ok := gr.groupLimiters[group]; ok {
		select {
		case <-ctx.Done():
			return time.Now()
		default:
			gr.masterLimit.Take()
		}

		select {
		case <-ctx.Done():
			return time.Now()
		default:
			return gl.Take()
		}
	}

	return time.Now()
}

func New(rate int, opts ...ratelimit.Option) GroupLimiter {
	return &groupLimiter{masterLimit: ratelimit.New(rate, opts...), groupLimiters: make(map[string]ratelimit.Limiter, 2)}
}

// AddGroup - add group to groupLimiter, method is not safe for concurrent use by multiple goroutines
func (gr *groupLimiter) AddGroup(groupName string, rate int, opts ...ratelimit.Option) GroupLimiter {
	gr.groupLimiters[groupName] = ratelimit.New(rate, opts...)

	return gr
}

// WithClock - allow set custom clock objects
func WithClock(clock clock.Clock) ratelimit.Option {
	return ratelimit.WithClock(clock)
}

// Per - allow configure time window for limits
func Per(per time.Duration) ratelimit.Option {
	return ratelimit.Per(per)
}

// WithSlack - allow collect unused requests for future, set how much unused requests can be collected
func WithSlack(slack int) ratelimit.Option {
	return ratelimit.WithSlack(slack)
}

// WithoutSlack - disable slack
func WithoutSlack() ratelimit.Option {
	return ratelimit.WithoutSlack
}
