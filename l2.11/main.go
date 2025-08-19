package main

import (
	"fmt"
	"sort"
	"strings"
)

// sortRunes сортирует символы в слове
func sortRunes(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	return string(runes)
}

// FindAnagramSets находит множества анаграмм по заданному словарю
func FindAnagramSets(words []string) map[string][]string {
	// ключ: буквы в нижнем регистре, отсортированы
	// значение: все слова в нижнем регистре
	anagramGroups := make(map[string][]string)

	// заполнение группы
	for _, w := range words {
		word := strings.ToLower(w)
		sorted := sortRunes(word)
		anagramGroups[sorted] = append(anagramGroups[sorted], word)
	}

	// итоговый результат
	result := make(map[string][]string)

	for _, group := range anagramGroups {
		if len(group) < 2 {
			continue // пропускаем одиночки
		}

		// уберём дубликаты
		unique := make(map[string]struct{})
		for _, w := range group {
			unique[w] = struct{}{}
		}

		// преобразуем обратно в срез
		final := make([]string, 0, len(unique))
		for w := range unique {
			final = append(final, w)
		}

		// сортируем по возрастанию
		sort.Strings(final)

		// ключ – первое слово в порядке появления (т.е. берем первый из group)
		result[group[0]] = final
	}

	return result
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	anagrams := FindAnagramSets(words)

	for key, group := range anagrams {
		fmt.Printf("%q: %v\n", key, group)
	}
}
