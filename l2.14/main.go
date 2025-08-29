package main

import (
	"fmt"
	"reflect"
	"time"
)

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}

func or(channels ...<-chan interface{}) <-chan interface{} {
	if len(channels) == 0 {
		c := make(chan interface{})
		close(c)
		return c
	}

	result := make(chan interface{})
	go func() {
		defer close(result)

		// преобразуем каналы в слайс SelectCase для reflect.Select
		var cases []reflect.SelectCase
		for _, ch := range channels {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ch),
			})
		}

		// ожидаем первый готовый канал
		_, _, _ = reflect.Select(cases)
	}()

	return result
}
