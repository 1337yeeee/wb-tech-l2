package main

import (
	"reflect"
	"sort"
	"testing"
)

// compareAnagramMaps сравнивает map[string][]string, так как порядок ключей не гарантирован
func compareAnagramMaps(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for keyA, valueA := range a {
		// Сортируем оба среза для сравнения
		sortedA := make([]string, len(valueA))
		copy(sortedA, valueA)
		sort.Strings(sortedA)

		valueB, exists := b[keyA]
		if !exists {
			return false
		}

		sortedB := make([]string, len(valueB))
		copy(sortedB, valueB)
		sort.Strings(sortedB)

		if !reflect.DeepEqual(sortedA, sortedB) {
			return false
		}
	}
	return true
}

// TestFindAnagramSets проверяем различные случаи
func TestFindAnagramSets(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string][]string
	}{
		{
			name:  "basic anagrams",
			input: []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик"},
			expected: map[string][]string{
				"пятак":  {"пятак", "пятка", "тяпка"},
				"листок": {"листок", "слиток", "столик"},
			},
		},
		{
			name:  "with duplicates",
			input: []string{"пятак", "пятка", "пятка", "тяпка", "тяпка"},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name:  "case insensitive",
			input: []string{"Пятак", "пЯтка", "Тяпка", "Листок", "Слиток"},
			expected: map[string][]string{
				"пятак":  {"пятак", "пятка", "тяпка"},
				"листок": {"листок", "слиток"},
			},
		},
		{
			name:  "single words are ignored",
			input: []string{"пятак", "стол", "стул", "пятка"},
			expected: map[string][]string{
				"пятак": {"пятак", "пятка"},
			},
		},
		{
			name:     "no anagrams",
			input:    []string{"стол", "стул", "окно"},
			expected: map[string][]string{},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: map[string][]string{},
		},
		{
			name:  "different length words",
			input: []string{"кот", "ток", "окот", "ктоо", "кто"},
			expected: map[string][]string{
				"кот":  {"кот", "кто", "ток"},
				"окот": {"ктоо", "окот"},
			},
		},
		{
			name:  "english anagrams",
			input: []string{"listen", "silent", "enlist", "google", "gogole"},
			expected: map[string][]string{
				"listen": {"enlist", "listen", "silent"},
				"google": {"google", "gogole"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindAnagramSets(tt.input)

			if !compareAnagramMaps(result, tt.expected) {
				t.Errorf("FindAnagramSets(%v) = %v, expected %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestFindAnagramSets_KeyIsFirstWord ключом должно быть первое слово "пятка" в нижнем регистре
func TestFindAnagramSets_KeyIsFirstWord(t *testing.T) {
	input := []string{"пятка", "пятак", "тяпка"}
	result := FindAnagramSets(input)

	expectedKey := "пятка"
	if _, exists := result[expectedKey]; !exists {
		t.Errorf("Expected key %q not found in result: %v", expectedKey, result)
	}
}

// TestFindAnagramSets_Sorting проверяем, что значения отсортированы
func TestFindAnagramSets_Sorting(t *testing.T) {
	input := []string{"тяпка", "пятка", "пятак"}
	result := FindAnagramSets(input)

	for key, group := range result {
		expectedSorted := make([]string, len(group))
		copy(expectedSorted, group)
		sort.Strings(expectedSorted)

		if !reflect.DeepEqual(group, expectedSorted) {
			t.Errorf("Group for key %s is not sorted: got %v, expected %v",
				key, group, expectedSorted)
		}
	}
}
