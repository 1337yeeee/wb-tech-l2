package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type fieldRange struct {
	start int
	end   int
}

type config struct {
	fields    []fieldRange
	delimiter string
	separated bool
}

func parseConfig() config {
	var fieldsStr string
	var delimiter string
	var separated bool

	flag.StringVar(&fieldsStr, "f", "", "fields to select (e.g., 1,3-5)")
	flag.StringVar(&delimiter, "d", "\t", "delimiter character")
	flag.BoolVar(&separated, "s", false, "only lines with delimiter")

	flag.Parse()

	fields, err := parseFields(fieldsStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var cfg = config{
		fields:    fields,
		delimiter: delimiter,
		separated: separated,
	}

	return cfg
}

func main() {
	cfg := parseConfig()

	args := flag.Args()

	if len(args) == 0 {
		// Если файлы не указаны, читаем из STDIN
		if err := cut(os.Stdin, os.Stdout, cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		// Обрабатываем каждый файл
		for _, filename := range args {
			if err := processFile(filename, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка обработки файла %s: %v\n", filename, err)
				os.Exit(1)
			}
		}
	}
}

func parseFields(fieldsStr string) ([]fieldRange, error) {
	var fields []fieldRange

	for _, part := range strings.Split(fieldsStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")

			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неправильный формат ввода: %s", part)
			}

			startStr := strings.TrimSpace(rangeParts[0])
			endStr := strings.TrimSpace(rangeParts[1])

			var start, end int
			var err error

			if startStr == "" {
				start = 1
			} else {
				start, err = strconv.Atoi(startStr)
				if err != nil || start < 1 {
					return nil, fmt.Errorf("неправильный формат числа: %s", startStr)
				}
			}

			if endStr == "" {
				end = -1
			} else {
				end, err = strconv.Atoi(endStr)
				if err != nil {
					return nil, fmt.Errorf("неправильный формат числа: %s", endStr)
				}
			}

			if end != -1 && start > end {
				return nil, fmt.Errorf("неправильный интервал: %d-%d", start, end)
			}

			fields = append(fields, fieldRange{
				start: start,
				end:   end,
			})
		} else {
			fieldNum, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || fieldNum < 1 {
				return nil, fmt.Errorf("неправильный формат числа: %s", part)
			}

			fields = append(fields, fieldRange{
				start: fieldNum,
				end:   fieldNum,
			})
		}
	}

	return fields, nil
}

func cut(input io.Reader, output io.Writer, cfg config) error {
	scanner := bufio.NewScanner(input)

	for scanner.Scan() {
		line := scanner.Text()

		if cfg.separated && !strings.Contains(line, cfg.delimiter) {
			continue
		}

		fields := strings.Split(line, cfg.delimiter)
		selectedFields := selectFields(fields, cfg.fields)

		fmt.Fprintln(output, strings.Join(selectedFields, cfg.delimiter))
	}

	return scanner.Err()
}

func selectFields(fields []string, fieldRanges []fieldRange) []string {
	var result []string
	fieldCount := len(fields)

	var selected = make(map[int]bool)

	// используем карту как множество и заполняем ee нужными индексами в соответствии с fieldRanges
	for _, fieldRange := range fieldRanges {
		start, end := fieldRange.start, fieldRange.end

		if end == -1 || end > fieldCount {
			end = fieldCount
		}

		for i := start; i <= end; i++ {
			if i <= fieldCount && !selected[i] {
				selected[i] = true
			}
		}
	}

	var idxCounter = 0
	var indices = make([]int, len(selected))
	for index := range selected {
		indices[idxCounter] = index
		idxCounter++
	}
	slices.Sort(indices)

	for _, index := range indices {
		if index <= fieldCount {
			result = append(result, fields[index-1])
		}
	}
	return result
}

func processFile(filename string, cfg config) error {
	// Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer file.Close()

	// Обрабатываем файл
	return cut(file, os.Stdout, cfg)
}
