package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseFields(t *testing.T) {
	tests := []struct {
		input    string
		expected []fieldRange
		hasError bool
	}{
		{
			input: "1,3-5",
			expected: []fieldRange{
				{start: 1, end: 1},
				{start: 3, end: 5},
			},
		},
		{
			input: "1-3,5",
			expected: []fieldRange{
				{start: 1, end: 3},
				{start: 5, end: 5},
			},
		},
		{
			input: "2-",
			expected: []fieldRange{
				{start: 2, end: -1},
			},
		},
		{
			input: "1,2,3",
			expected: []fieldRange{
				{start: 1, end: 1},
				{start: 2, end: 2},
				{start: 3, end: 3},
			},
		},
		{
			input:    "0",
			hasError: true,
		},
		{
			input: "1-",
			expected: []fieldRange{
				{start: 1, end: -1},
			},
			hasError: false,
		},
	}

	for _, test := range tests {
		result, err := parseFields(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
			continue
		}

		if len(result) != len(test.expected) {
			t.Errorf("For input %s, expected %d ranges, got %d", test.input, len(test.expected), len(result))
			continue
		}

		for i, expected := range test.expected {
			if result[i].start != expected.start || result[i].end != expected.end {
				t.Errorf("For input %s, range %d: expected {%d,%d}, got {%d,%d}",
					test.input, i, expected.start, expected.end, result[i].start, result[i].end)
			}
		}
	}
}

func TestCutBasic(t *testing.T) {
	input := "a\tb\tc\td\te\nf\tg\th\ti\tj\n"
	expected := "a\tc\td\nf\th\ti\n"

	cfg := config{
		fields:    []fieldRange{{start: 1, end: 1}, {start: 3, end: 4}},
		delimiter: "\t",
		separated: false,
	}

	var output bytes.Buffer
	reader := strings.NewReader(input)

	err := cut(reader, &output, cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestCutSeparatedFlag(t *testing.T) {
	input := "a\tb\tc\nd\te\nno_tabs_here\nf\tg\th\n"
	expected := "a\tc\nd\nf\th\n"

	cfg := config{
		fields:    []fieldRange{{start: 1, end: 1}, {start: 3, end: 3}},
		delimiter: "\t",
		separated: true,
	}

	var output bytes.Buffer
	reader := strings.NewReader(input)

	err := cut(reader, &output, cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output.String() != expected {
		t.Errorf("Expected %q, got %q", expected, output.String())
	}
}

func TestSelectFields(t *testing.T) {
	allFields := []string{"a", "b", "c", "d", "e"}

	tests := []struct {
		ranges   []fieldRange
		expected []string
	}{
		{
			ranges:   []fieldRange{{start: 1, end: 1}, {start: 3, end: 4}},
			expected: []string{"a", "c", "d"},
		},
		{
			ranges:   []fieldRange{{start: 2, end: 2}, {start: 1, end: 1}},
			expected: []string{"a", "b"},
		},
		{
			ranges:   []fieldRange{{start: 3, end: 10}}, // out of bounds
			expected: []string{"c", "d", "e"},
		},
	}

	for _, test := range tests {
		result := selectFields(allFields, test.ranges)
		if len(result) != len(test.expected) {
			t.Errorf("Expected length %d, got %d", len(test.expected), len(result))
			continue
		}

		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("Expected %s, got %s", expected, result[i])
			}
		}
	}
}
