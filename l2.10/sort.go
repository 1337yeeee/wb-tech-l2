package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// monthMap используется для сортировки по месяцам (-M)
var monthMap = map[string]int{
	"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4,
	"May": 5, "Jun": 6, "Jul": 7, "Aug": 8,
	"Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
}

var (
	colFlag   = flag.Int("k", 0, "sort by column N (1-based, default: whole line)")
	numFlag   = flag.Bool("n", false, "compare by numeric value")
	reverse   = flag.Bool("r", false, "reverse the result")
	unique    = flag.Bool("u", false, "output only unique lines")
	monthFlag = flag.Bool("M", false, "compare by month name")
	ignoreBl  = flag.Bool("b", false, "ignore trailing blanks")
	check     = flag.Bool("c", false, "check if input is sorted")
	human     = flag.Bool("h", false, "compare human-readable numbers (1K 2M ...)")
)

func readLines(lines []string) []string {
	if len(flag.Args()) > 0 {
		for _, fname := range flag.Args() {
			f, err := os.Open(fname)
			if err != nil {
				fmt.Fprintln(os.Stderr, "cannot open file:", err)
				os.Exit(1)
			}
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if *ignoreBl {
					line = strings.TrimRight(line, " ")
				}
				lines = append(lines, line)
			}
			errClose := f.Close()
			if errClose != nil {
				fmt.Fprintln(os.Stderr, "cannot close file:", errClose)
				os.Exit(1)
			}
		}
	} else {
		// Иначе читаем из STDIN
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if *ignoreBl {
				line = strings.TrimRight(line, " ")
			}
			lines = append(lines, line)
		}
	}

	return lines
}

// extractKey извлекает значение из строки для сортировки
func extractKey(line string, col int) string {
	if col <= 0 {
		return line
	}
	parts := strings.Split(line, "\t")
	if col-1 < len(parts) {
		return parts[col-1]
	}
	return line
}

// parseHuman читает строку вида "10K", "2M"
func parseHuman(s string) float64 {
	mult := 1.0
	if len(s) == 0 {
		return 0
	}
	last := s[len(s)-1]
	num := s
	switch last {
	case 'K', 'k':
		mult = 1024
		num = s[:len(s)-1]
	case 'M', 'm':
		mult = 1024 * 1024
		num = s[:len(s)-1]
	case 'G', 'g':
		mult = 1024 * 1024 * 1024
		num = s[:len(s)-1]
	}
	v, _ := strconv.ParseFloat(num, 64)
	return v * mult
}

func lessSort(lines []string) func(i, j int) bool {
	return func(i, j int) bool {
		a := extractKey(lines[i], *colFlag)
		b := extractKey(lines[j], *colFlag)

		var less bool

		if *numFlag {
			// сортировка чисел
			af, _ := strconv.ParseFloat(a, 64)
			bf, _ := strconv.ParseFloat(b, 64)
			if af != bf {
				less = af < bf
			} else {
				less = a < b // если полученные значения равны, сравниваем строки (например, 1.0 == 1.00)
			}
		} else if *monthFlag {
			// сортировка по месяцу
			am, aok := monthMap[a]
			bm, bok := monthMap[b]
			if aok && bok && am != bm {
				less = am < bm
			} else {
				less = a < b // если месяц не опознан или равен
			}
		} else if *human {
			// человекочитаемые размеры
			af := parseHuman(a)
			bf := parseHuman(b)
			if af != bf {
				less = af < bf
			} else {
				less = a < b // если полученные значения равны, сравниваем строки (например, 1K == 1024)
			}
		} else {
			// сравнение строк
			less = a < b
		}

		if *reverse {
			return !less
		}
		return less
	}
}

// checkSorted проверяет, отсортированы ли передаваемые значения
func checkSorted(lines []string) error {
	less := lessSort(lines)
	for i := 1; i < len(lines); i++ {
		// если порядок нарушен (строка i < предыдущей)
		if less(i, i-1) {
			// выводим первую "неправильную" строку
			return fmt.Errorf("sort: not sorted: %s", lines[i])
		}
	}
	return nil
}

func main() {
	flag.Parse()

	var lines []string
	lines = readLines(lines)

	if *check {
		err := checkSorted(lines)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else {
		sort.Slice(lines, lessSort(lines))
	}

	// unique
	if *unique {
		var uniq []string
		var prev string
		for _, l := range lines {
			if l != prev {
				uniq = append(uniq, l)
				prev = l
			}
		}
		lines = uniq
	}

	for _, l := range lines {
		fmt.Println(l)
	}
}
