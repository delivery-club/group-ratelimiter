package ratelimiter

import (
	"go.uber.org/ratelimit"
)

// Приоритизация групп возможна засчет разницы между их лимитами,
// чем выше лимит группы, тем выше ее приоритет,
// тем не менее общее количество запросов будет ограниченна основным лимитом (MasterLimit)
type groupLimitConfig interface {
	MasterRate() int
	GroupRates() map[string]int
}

type GroupLimiter interface {
	Take(groupName string)
}

// groupLimiter описание условий:
// 1. есть лимиты группы
// 2. есть общий лимит
// 3. группа должна уметь использовать часть лимитов общих, с возможностью приоритезации
// 4. TODO рассмотреть возможность автонаращивание лимитов группами при неиспользовании их одной из групп
type groupLimiter map[string]*dependentLimiter

type dependentLimiter struct {
	masterLimit ratelimit.Limiter
	groupLimit  ratelimit.Limiter
}

func (rl groupLimiter) Take(group string) {
	if gl, ok := rl[group]; ok {
		gl.groupLimit.Take()
		gl.masterLimit.Take()
	}
}

func NewRateLimiter(config groupLimitConfig, opts ...ratelimit.Option) GroupLimiter {
	groupLimiter := make(groupLimiter, len(config.GroupRates()))
	mainLimiter := ratelimit.New(config.MasterRate(), ratelimit.WithoutSlack)

	for group, rate := range config.GroupRates() {
		groupLimiter[group] = &dependentLimiter{
			masterLimit: mainLimiter,
			groupLimit:  ratelimit.New(rate, opts...),
		}
	}

	return groupLimiter
}
