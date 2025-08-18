package main

import (
	"errors"
	"fmt"
	"strings"
)

// BaseState режим по умолчанию для StringUnpacking
const (
	BaseState   = iota // Обычный режим работы
	DigitState         // Режим работы, когда была встречена цифра
	EscapeState        // Режим работы, когда был встречен символ экранирования
)

// ErrInvalidString ошибка возвращается, когда строка невалидная
var ErrInvalidString = errors.New("invalid string")

// ErrDanglingEscape ошибка возвращается, когда на конце строки остается висящий символ экранирования
var ErrDanglingEscape = errors.New("dangling escape at end")

// StringUnpacking распаковывает строку, содержащую повторяющиеся символы
func StringUnpacking(str string) (string, error) {
	var builder strings.Builder
	var prev string
	var state = BaseState
	var hasLetter = false
	var char rune
	var n = 0

	flushDigit := func() error {
		if prev == "" {
			return ErrInvalidString
		}

		builder.WriteString(strings.Repeat(prev, n))
		prev = ""
		n = 0
		return nil
	}

	flushPrev := func() {
		builder.WriteString(prev)
		prev = string(char)
	}

	for _, char = range str {
		if char >= '0' && char <= '9' {
			switch state {
			case BaseState:
				n = n*10 + int(char-'0')
				state = DigitState
			case DigitState:
				n = n*10 + int(char-'0')
			case EscapeState:
				flushPrev()
				state = BaseState
			}
		} else if char == '\\' {
			switch state {
			case BaseState:
				state = EscapeState
			case DigitState:
				if err := flushDigit(); err != nil {
					return "", err
				}
				state = EscapeState
			case EscapeState:
				flushPrev()
				state = BaseState
			}
		} else {
			switch state {
			case BaseState:
				flushPrev()
				hasLetter = true
			case DigitState:
				if err := flushDigit(); err != nil {
					return "", err
				}
				prev = string(char)
				state = BaseState
				hasLetter = true
			case EscapeState:
				flushPrev()
				state = BaseState
				hasLetter = true
			}
		}
	}

	switch state {
	case DigitState:
		if err := flushDigit(); err != nil {
			return "", err
		}
	case EscapeState:
		return "", ErrDanglingEscape
	default:
		builder.WriteString(prev)
	}

	if !hasLetter && builder.Len() == 0 && len(str) != 0 {
		return "", ErrInvalidString
	}

	return builder.String(), nil
}

func main() {
	var tests = []struct {
		input  string
		output string
		error  error
	}{
		{"qwe\\45", "qwe44444", nil},
		{"a4bc2d5e", "aaaabccddddde", nil},
		{"abcd", "abcd", nil},
		{"45", "", ErrInvalidString},
		{"", "", nil},
		{"qwe\\4\\5", "qwe45", nil},
		{"a12", "aaaaaaaaaaaa", nil},
		{"a0", "", nil},
		{"ab0c", "ac", nil},
		{"a\\12", "a11", nil},
		{"\\3", "3", nil},
		{"\\\\", "\\", nil},
		{"a\\\\3", "a\\\\\\", nil},
		{"😀3", "😀😀😀", nil},
		{"4a", "", ErrInvalidString},
		{"4", "", ErrInvalidString},
		{"abc\\", "", ErrDanglingEscape},
	}

	count := 0
	fail := 0
	success := 0
	for i, test := range tests {
		if i == 4 {
			i = 4
		}
		count++
		result, err := StringUnpacking(test.input)
		fmt.Printf("Test %d:\n", i)
		fmt.Printf("input:\t%s\n", test.input)
		fmt.Printf("output:\t%s\n", test.output)
		fmt.Printf("result:\t%s\n", result)
		if err != nil {
			fmt.Printf("Error:\t%s\n", err.Error())
		}

		isSuccess := result == test.output && err == test.error
		if isSuccess {
			success++
		} else {
			fail++
		}
		fmt.Printf("Is OK: %v\n\n", isSuccess)
	}

	fmt.Printf("%d tests were ran, %d successed, %d failed\n", count, success, fail)
}

//Test 0:
//input:	qwe\45
//output:	qwe44444
//result:	qwe44444
//Is OK: true
//
//Test 1:
//input:	a4bc2d5e
//output:	aaaabccddddde
//result:	aaaabccddddde
//Is OK: true
//
//Test 2:
//input:	abcd
//output:	abcd
//result:	abcd
//Is OK: true
//
//Test 3:
//input:	45
//output:
//result:
//Error:	invalid string
//Is OK: true
//
//Test 4:
//input:
//output:
//result:
//Is OK: true
//
//Test 5:
//input:	qwe\4\5
//output:	qwe45
//result:	qwe45
//Is OK: true
//
//Test 6:
//input:	a12
//output:	aaaaaaaaaaaa
//result:	aaaaaaaaaaaa
//Is OK: true
//
//Test 7:
//input:	a0
//output:
//result:
//Is OK: true
//
//Test 8:
//input:	ab0c
//output:	ac
//result:	ac
//Is OK: true
//
//Test 9:
//input:	a\12
//output:	a11
//result:	a11
//Is OK: true
//
//Test 10:
//input:	\3
//output:	3
//result:	3
//Is OK: true
//
//Test 11:
//input:	\\
//output:	\
//result:	\
//Is OK: true
//
//Test 12:
//input:	a\\3
//output:	a\\\
//result:	a\\\
//Is OK: true
//
//Test 13:
//input:	😀3
//output:	😀😀😀
//result:	😀😀😀
//Is OK: true
//
//Test 14:
//input:	4a
//output:
//result:
//Error:	invalid string
//Is OK: true
//
//Test 15:
//input:	4
//output:
//result:
//Error:	invalid string
//Is OK: true
//
//Test 16:
//input:	abc\
//output:
//result:
//Error:	dangling escape at end
//Is OK: true
//
//17 tests were ran, 17 successed, 0 failed
