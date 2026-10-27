package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tokenBucket := NewNaiveTokenBucketWithChannels(5, 1*time.Second)
	tokenBucket.Start(ctx)
	tokenBucket.Start(ctx)
	wg := &sync.WaitGroup{}
	for i := range 20 {
		wg.Go(func() {
			time.Sleep(time.Duration(rand.IntN(100)) * 100 * time.Millisecond)
			if tokenBucket.Allow(ctx) {
				fmt.Printf("%d: разрешено (осталось %d)\n", i, tokenBucket.Tokens())
			} else {
				fmt.Printf("%d: отклонено (осталось %d)\n", i, tokenBucket.Tokens())
			}
		})
	}
	wg.Wait()
}

// учебная роеализация
type NaiveTokenBucketWithChannels struct {
	tokens  chan struct{}
	ticker  *time.Ticker
	started atomic.Bool
}

func NewNaiveTokenBucketWithChannels(capacity int, refillRate time.Duration) *NaiveTokenBucketWithChannels {
	tokenBucket := &NaiveTokenBucketWithChannels{
		tokens: make(chan struct{}, capacity),
		ticker: time.NewTicker(refillRate),
	}
	for range capacity {
		tokenBucket.tokens <- struct{}{}
	}
	return tokenBucket
}

func (tb *NaiveTokenBucketWithChannels) Start(ctx context.Context) {
	if !tb.started.CompareAndSwap(false, true) {
		fmt.Println("уже запущен")
	}
	go func() {
		defer tb.ticker.Stop()
		for {
			select {
			case <-tb.ticker.C:
				select {
				case tb.tokens <- struct{}{}:
					fmt.Println("Добавлен 1 токен")
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (tb *NaiveTokenBucketWithChannels) Allow(ctx context.Context) bool {
	// Первый select: ctx ещё не отменён → идём дальше.
	// Контекст отменяется (другая горутина вызвала cancel()).
	// Второй select: tb.tokens готов → возвращаем true.
	// Результат: Allow вернул true, хотя контекст уже отменён.
	// Это не атомарно и может привести к ошибкам: можно пропустить горутину,
	// которая не успеет выполнится и может быть потеря консистентности.
	select {
	case <-ctx.Done():
		return false
	default:
	}

	select {
	case <-tb.tokens:
		return true
	default:
		return false
	}
}

func (tb *NaiveTokenBucketWithChannels) Tokens() int {
	// не стоит полагаться на вывод. если нужен точный результат для принятия решения -
	// нужен mutex.
	return len(tb.tokens)
}
