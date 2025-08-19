package main

import (
	"sort"
	"testing"
)

func resetFlags() {
	*numFlag = false
	*reverse = false
	*unique = false
	*monthFlag = false
	*ignoreBl = false
	*check = false
	*human = false
	*colFlag = 0
}

// TestExtractKey проверяет функцию extractKey
func TestExtractKey(t *testing.T) {
	line := "foo\tbar\tbaz"

	tests := []struct {
		col  int
		want string
	}{
		{0, line},  // без указания колонки
		{1, "foo"}, // первая колонка
		{2, "bar"}, // вторая колонка
		{3, "baz"}, // третья колонка
		{4, line},  // колонки нет
		{-1, line}, // невалидный col
	}

	for _, tt := range tests {
		got := extractKey(line, tt.col)
		if got != tt.want {
			t.Errorf("extractKey(%q, %d) = %q, want %q", line, tt.col, got, tt.want)
		}
	}
}

// TestParseHuman проверяет функцию parseHuman
func TestParseHuman(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"10", 10},
		{"1K", 1024},
		{"2M", 2 * 1024 * 1024},
		{"3G", 3 * 1024 * 1024 * 1024},
		{"5k", 5 * 1024},
		{"7m", 7 * 1024 * 1024},
	}

	for _, tt := range tests {
		got := parseHuman(tt.in)
		if got != tt.want {
			t.Errorf("parseHuman(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// TestLessSortNumeric проверяет функцию lessSort с флагом -n
func TestLessSortNumeric(t *testing.T) {
	lines := []string{"10", "2", "30"}

	resetFlags()
	*numFlag = true

	sort.Slice(lines, lessSort(lines))

	want := []string{"2", "10", "30"}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("numeric sort mismatch: got %v, want %v", lines, want)
			break
		}
	}
}

// TestLessSortReverse проверяет функцию lessSort с флагом -r
func TestLessSortReverse(t *testing.T) {
	lines := []string{"a", "c", "b"}

	resetFlags()
	*reverse = true

	sort.Slice(lines, lessSort(lines))

	want := []string{"c", "b", "a"}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("reverse sort mismatch: got %v, want %v", lines, want)
			break
		}
	}
}

// TestLessSortReverse проверяет функцию lessSort с флагом -M
func TestLessSortMonth(t *testing.T) {
	// последнее значение ("Feb ") с дополнительным пробелом, должно идти после значения без пробела
	lines := []string{"Mar", "Feb", "Jan", "Feb "}

	resetFlags()
	*monthFlag = true

	sort.Slice(lines, lessSort(lines))

	want := []string{"Jan", "Feb", "Feb ", "Mar"}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("month sort mismatch: got %v, want %v", lines, want)
			break
		}
	}
}

// TestLessSortReverse проверяет функцию lessSort с флагом -h
func TestLessSortHuman(t *testing.T) {
	lines := []string{"2K", "1M", "2048", "512"}

	resetFlags()
	*human = true

	sort.Slice(lines, lessSort(lines))

	want := []string{"512", "2048", "2K", "1M"}

	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("human sort mismatch: got %v, want %v", lines, want)
			break
		}
	}
}

// TestCheckSorted_StringSorted проверяет отсортированные строки с флагом -c
func TestCheckSorted_StringSorted(t *testing.T) {
	resetFlags()
	lines := []string{"a", "b", "c"}
	if err := checkSorted(lines); err != nil {
		t.Errorf("expected sorted, got error %v", err)
	}
}

// TestCheckSorted_StringSorted проверяет НЕ отсортированные строки с флагом -c
func TestCheckSorted_StringUnsorted(t *testing.T) {
	resetFlags()
	lines := []string{"b", "a"}
	if err := checkSorted(lines); err == nil {
		t.Errorf("expected error for unsorted input")
	}
}

// TestCheckSorted_StringSorted проверяет отсортированные строки с флагом -cn
func TestCheckSorted_Numeric(t *testing.T) {
	resetFlags()
	*numFlag = true
	lines := []string{"1", "2", "10"}
	if err := checkSorted(lines); err != nil {
		t.Errorf("expected sorted numbers, got error %v", err)
	}
}

// TestCheckSorted_StringSorted проверяет НЕ отсортированные строки с флагом -cn
func TestCheckSorted_NumericUnsorted(t *testing.T) {
	resetFlags()
	*numFlag = true
	lines := []string{"10", "2", "1"}
	if err := checkSorted(lines); err == nil {
		t.Errorf("expected error for unsorted numeric input")
	}
}

// TestCheckSorted_StringSorted проверяет отсортированные строки с флагом -cM
func TestCheckSorted_Months(t *testing.T) {
	resetFlags()
	*monthFlag = true
	lines := []string{"Jan", "Feb", "Mar"}
	if err := checkSorted(lines); err != nil {
		t.Errorf("expected sorted months, got error %v", err)
	}
}

// TestCheckSorted_StringSorted проверяет НЕ отсортированные строки с флагом -cM
func TestCheckSorted_MonthsUnsorted(t *testing.T) {
	resetFlags()
	*monthFlag = true
	lines := []string{"Mar", "Jan"}
	if err := checkSorted(lines); err == nil {
		t.Errorf("expected error for unsorted months")
	}
}

// TestCheckSorted_StringSorted проверяет отсортированные строки с флагом -ch
func TestCheckSorted_HumanReadable(t *testing.T) {
	resetFlags()
	*human = true
	lines := []string{"1K", "2K", "10M"}
	if err := checkSorted(lines); err != nil {
		t.Errorf("expected sorted human-readable, got error %v", err)
	}
}

// TestCheckSorted_StringSorted проверяет НЕ отсортированные строки с флагом -ch
func TestCheckSorted_HumanReadableUnsorted(t *testing.T) {
	resetFlags()
	*human = true
	lines := []string{"10M", "2K"}
	if err := checkSorted(lines); err == nil {
		t.Errorf("expected error for unsorted human-readable numbers")
	}
}

// TestCheckSorted_StringSorted проверяет НЕ отсортированные строки с флагом -cr
func TestCheckSorted_Reverse(t *testing.T) {
	resetFlags()
	*reverse = true
	lines := []string{"c", "b", "a"}
	if err := checkSorted(lines); err != nil {
		t.Errorf("expected sorted in reverse, got error %v", err)
	}
}
