package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	tokenBucket := NewTokenBucketWithMutex(5.0, 1)
	wg := &sync.WaitGroup{}
	for i := range 20 {
		wg.Go(func() {
			time.Sleep(time.Duration(rand.IntN(200)) * 50 * time.Millisecond)
			if tokenBucket.Allow(1.0) {
				fmt.Printf("%d: разрешено (осталось ~%.2f)\n", i, tokenBucket.Tokens())
			} else {
				fmt.Printf("%d: отклонено (осталось ~%.2f)\n", i, tokenBucket.Tokens())
			}
		})
	}
	wg.Wait()
}

type NaiveTokenBucketWithMutex struct {
	capacity   float64
	refillRate float64
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucketWithMutex(capacity, refillRate float64) *NaiveTokenBucketWithMutex {
	return &NaiveTokenBucketWithMutex{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (tb *NaiveTokenBucketWithMutex) Allow(cost float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens < cost {
		return false
	}
	tb.tokens -= cost
	return true
}

func (tb *NaiveTokenBucketWithMutex) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

func (tb *NaiveTokenBucketWithMutex) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+tb.refillRate*elapsed)
	tb.lastRefill = now
}
