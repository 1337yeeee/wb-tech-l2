package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const Divider = "--"

type MatchResult struct {
	line    string
	lineNum int
	matched bool
	printed bool
}

var (
	afterContext  = flag.Int("A", 0, "Print N lines after each match")
	beforeContext = flag.Int("B", 0, "Print N lines before each match")
	context       = flag.Int("C", 0, "Print N lines of context around each match")
	countOnly     = flag.Bool("c", false, "Print only count of matching lines")
	ignoreCase    = flag.Bool("i", false, "Ignore case")
	invertMatch   = flag.Bool("v", false, "Invert match")
	fixedString   = flag.Bool("F", false, "Treat pattern as fixed string")
	lineNumber    = flag.Bool("n", false, "Print line numbers")
	extended      = flag.Bool("E", false, "Extended mode, does nothing, for grep compatibility")
)

var pattern string

func filterInput(input *os.File) error {
	scanner := bufio.NewScanner(input)
	var lines []MatchResult
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		matched := isMatch(line)
		lines = append(lines, MatchResult{
			line:    line,
			lineNum: lineNum,
			matched: matched,
			printed: false,
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Если нужно только количество
	if *countOnly {
		count := 0
		for _, result := range lines {
			if result.matched {
				count++
			}
		}
		fmt.Println(count)
		return nil
	}

	// Обработка вывода
	return processOutput(lines)
}

func isMatch(line string) bool {
	var matched bool
	searchLine := line
	searchPattern := pattern

	if *ignoreCase {
		searchLine = strings.ToLower(searchLine)
		searchPattern = strings.ToLower(searchPattern)
	}

	if *fixedString {
		matched = strings.Contains(searchLine, searchPattern)
	} else {
		// Компилируем регулярное выражение с учетом регистра
		var re *regexp.Regexp
		var err error

		if *ignoreCase {
			re, err = regexp.Compile("(?i)" + pattern)
		} else {
			re, err = regexp.Compile(pattern)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid regex pattern: %v\n", err)
			os.Exit(2)
		}
		matched = re.MatchString(line)
	}

	if *invertMatch {
		return !matched
	}
	return matched
}

// processOutput печатает результат
func processOutput(lines []MatchResult) error {
	// Сначала помечаем все строки для печати
	for i, result := range lines {
		if result.matched {
			// Помечаем строки до
			start := max(0, i-*beforeContext)
			for j := start; j < i; j++ {
				lines[j].printed = true
			}

			// Помечаем саму строку
			lines[i].printed = true

			// Помечаем строки после
			end := min(len(lines)-1, i+*afterContext)
			for j := i + 1; j <= end; j++ {
				lines[j].printed = true
			}
		}
	}

	lastPrinted := -2 // Инициализируем значением, гарантирующим, что первая строка не будет иметь разделитель
	hasContext := *afterContext > 0 || *beforeContext > 0 || *context > 0

	for i, result := range lines {
		if result.printed {
			// Добавляем разделитель только если есть контекст И это не первая печатаемая строка
			// И предыдущая строка не была напечатана (есть разрыв)
			if hasContext && lastPrinted != -2 && lastPrinted != i-1 {
				fmt.Println(Divider)
			}

			if *lineNumber {
				fmt.Printf("%d:", result.lineNum)
			}
			fmt.Println(result.line)
			lastPrinted = i
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if *context > 0 {
		*afterContext = *context
		*beforeContext = *context
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: grep [flags] pattern [file]")
		os.Exit(1)
	}

	pattern = args[0]
	var filename string
	if len(args) > 1 {
		filename = args[1]
	}

	var input *os.File
	if filename == "" {
		input = os.Stdin
	} else {
		var err error
		input, err = os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer input.Close()
	}

	err := filterInput(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
