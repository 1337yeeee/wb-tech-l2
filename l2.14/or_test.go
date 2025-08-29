package main

import (
	"testing"
	"time"
)

// Вспомогательная функция для создания сигнального канала с задержкой
func sig(after time.Duration) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}

// Тест: or() без аргументов → закрытый канал
func TestOrZeroChannels(t *testing.T) {
	merged := or()
	select {
	case <-merged:
		// OK: канал уже закрыт
	default:
		t.Error("ожидался закрытый канал, но он блокирует")
	}
}

// Тест: один канал — возвращается он сам (по сути)
func TestOrOneChannel(t *testing.T) {
	c := sig(10 * time.Millisecond)
	start := time.Now()
	<-or(c)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Error("канал закрылся слишком быстро")
	}
}

// Тест: несколько каналов, самый быстрый определяет результат
func TestOrMultipleChannels(t *testing.T) {
	start := time.Now()
	<-or(
		sig(100*time.Millisecond),
		sig(50*time.Millisecond),
		sig(200*time.Millisecond),
		sig(10*time.Millisecond), // самый быстрый
	)
	elapsed := time.Since(start)

	if elapsed > 15*time.Millisecond {
		t.Errorf("ожидалось завершение ~10ms, но прошло %v", elapsed)
	}
}

// Тест: два канала, один уже закрыт — or должен завершиться мгновенно
func TestOrWithAlreadyClosedChannel(t *testing.T) {
	closedChan := make(chan interface{})
	close(closedChan)

	start := time.Now()
	<-or(
		sig(100*time.Millisecond),
		closedChan,
	)
	elapsed := time.Since(start)

	if elapsed > 5*time.Millisecond {
		t.Errorf("ожидалось мгновенное завершение, но прошло %v", elapsed)
	}
}

// Тест: два канала, оба уже закрыты — or завершается мгновенно
func TestOrWithAllClosedChannels(t *testing.T) {
	c1 := make(chan interface{})
	c2 := make(chan interface{})
	close(c1)
	close(c2)

	start := time.Now()
	<-or(c1, c2)
	elapsed := time.Since(start)

	if elapsed > 5*time.Millisecond {
		t.Errorf("ожидалось мгновенное завершение, но прошло %v", elapsed)
	}
}
