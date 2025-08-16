package main

import (
	"testing"
	"time"
)

// Test: успешное получение времени с публичного пула NTP
func TestGetNTPTimeSuccess(t *testing.T) {
	server := "pool.ntp.org"
	var ntpTimeProvider TimeProvider = NTPProvider{}
	got, err := ntpTimeProvider.Now(server)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.IsZero() {
		t.Fatalf("expected non-zero time, got %v", got)
	}

	if diff := time.Since(got); diff > 3*time.Second || diff < -3*time.Second {
		t.Errorf("time difference too large: %v", diff)
	}
}

// Test: неправильный адрес NTP-сервера, должен вернуть ошибку
func TestGetNTPTimeInvalidServer(t *testing.T) {
	server := "invalid.ntp.server"
	var ntpTimeProvider TimeProvider = NTPProvider{}
	_, err := ntpTimeProvider.Now(server)
	if err == nil {
		t.Fatalf("expected error for invalid server, got nil")
	}
}
