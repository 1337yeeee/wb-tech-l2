package main

import (
	"errors"
	"testing"
	"time"
)

// FakeProvider — тестовая реализация TimeProvider.
type FakeProvider struct {
	TimeToReturn time.Time
	ErrToReturn  error
}

func (f FakeProvider) Now(server string) (time.Time, error) {
	if f.ErrToReturn != nil {
		return time.Time{}, f.ErrToReturn
	}
	return f.TimeToReturn, nil
}

// Тест: успешное получение времени
func TestTimeProviderSuccess(t *testing.T) {
	expected := time.Date(2025, 8, 16, 12, 0, 0, 0, time.UTC)
	provider := FakeProvider{TimeToReturn: expected}

	got, err := provider.Now("fake.ntp.server")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !got.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

// Тест: ошибка от провайдера
func TestTimeProviderError(t *testing.T) {
	provider := FakeProvider{ErrToReturn: errors.New("network unreachable")}

	_, err := provider.Now("fake.ntp.server")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
