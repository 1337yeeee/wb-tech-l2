#!/usr/bin/env bash
set -Eeuo pipefail

# Сборка бинаря
go build -o mygrep main.go

# Массив тестов: "<name>|<pattern>|<file>|<flags>
TESTS=(
"basic|test|testcases/test1.txt"
"ignore_case|test|testcases/test1.txt|-i"
"context_after|test|testcases/test1.txt|-A 1"
"context_before|test|testcases/test1.txt|-B 1"
"context_around|test|testcases/test1.txt|-C 1"
"count_only|test|testcases/test1.txt|-c"
"invert_match|test|testcases/test1.txt|-v"
"line_numbers|test|testcases/test1.txt|-n"
"fixed_string|error|testcases/test2.txt|-F"
"regex_pattern|error [0-9]+|testcases/test2.txt|-E"
"stdin|test|testcases/test1.txt"  # будет обработан отдельно

# Комбинированные тесты
"combo_ci|test|testcases/test1.txt|-C 1 -i"
"combo_nv|test|testcases/test1.txt|-n -v"
"regex_invert|[0-9]|testcases/test2.txt|-v"

# Тесты с игнорированием регистра для regex
"regex_ignore_case|ERROR|testcases/test2.txt|-i"
"regex_case_sensitive|ERROR|testcases/test2.txt"  # без -i

# Тесты с граничными случаями
"empty_file|test|testcases/empty.txt"
"no_matches|nonexistent|testcases/test1.txt"

# Тесты с ошибками (должны завершаться с ошибкой у обоих)
"error_invalid_regex|[|testcases/test1.txt"
"error_file_not_found|test|nonexistent.txt"

# Тесты с повторяющимися строками
"duplicates_basic|cherry|testcases/test3.txt"
"duplicates_ignore_case|cherry|testcases/test3.txt|-i"
"duplicates_count|cherry|testcases/test3.txt|-c"
"duplicates_line_numbers|cherry|testcases/test3.txt|-n"
"duplicates_context|cherry|testcases/test3.txt|-A 1 -B 1"
"duplicates_invert|cherry|testcases/test3.txt|-v"
"duplicates_fixed|cherry|testcases/test3.txt|-F"
"duplicates_case_sensitive|Cherry|testcases/test3.txt"
"duplicates_elderberry|Elderberry|testcases/test3.txt"
"duplicates_date|date|testcases/test3.txt|-i"
)

ok=0
fail=0
tmp_mygrep="$(mktemp)"
tmp_grep="$(mktemp)"

# Создаем пустой файл для тестов
touch testcases/empty.txt

echo "Starting grep comparison tests..."
echo "================================="

for t in "${TESTS[@]}"; do
    IFS='|' read -r name pattern file grep_flags mygrep_flags <<< "$t"

    # Если не указаны отдельные флаги, используем одинаковые
    if [[ -z "$mygrep_flags" ]]; then
        mygrep_flags="$grep_flags"
    fi

    # Формируем команды
    if [[ "$name" == "stdin" ]]; then
        mygrep_cmd="cat '$file' | ./mygrep $mygrep_flags '$pattern'"
        grep_cmd="cat '$file' | grep $grep_flags '$pattern'"
    else
        mygrep_cmd="./mygrep $mygrep_flags '$pattern' '$file'"
        grep_cmd="grep $grep_flags '$pattern' '$file'"
    fi

    echo "▶️  $name: ./mygrep $mygrep_flags '$pattern' $file"

    # Проверяем статусы завершения
    if [[ "$name" == error_* ]]; then
        set +e
        # Запускаем обе команды и сохраняем вывод
        bash -c "$mygrep_cmd" > "$tmp_mygrep" 2>/dev/null
        mygrep_exit=$?

        bash -c "$grep_cmd" > "$tmp_grep" 2>/dev/null
        grep_exit=$?
        set -e

        # Для тестов с ошибками оба должны завершиться с ошибкой
        if [[ $mygrep_exit -ne 0 && $grep_exit -ne 0 ]]; then
            echo "✅  $name passed (both failed as expected)"
            ((ok++))
        else
            echo "❌  $name failed (exit codes differ: mygrep=$mygrep_exit, grep=$grep_exit)"
            ((fail++))
        fi
    else
        # Запускаем обе команды и сохраняем вывод
        bash -c "$mygrep_cmd" > "$tmp_mygrep" 2>/dev/null || true
        mygrep_exit=$?

        bash -c "$grep_cmd" > "$tmp_grep" 2>/dev/null || true
        grep_exit=$?

        # Для обычных тестов проверяем и вывод, и статус
        if [[ $mygrep_exit -eq $grep_exit ]]; then
            if diff -u "$tmp_mygrep" "$tmp_grep" >/dev/null; then
                echo "✅  $name passed (output and exit code match)"
                ((ok++))
            else
                echo "❌  $name failed (output differs but exit codes match)"
                echo "=== mygrep output ==="
                cat "$tmp_mygrep"
                echo "=== grep output ==="
                cat "$tmp_grep"
                echo "==================="
                ((fail++))
            fi
        else
            echo "❌  $name failed (exit codes differ: mygrep=$mygrep_exit, grep=$grep_exit)"
            echo "=== mygrep output ==="
            cat "$tmp_mygrep"
            echo "=== grep output ==="
            cat "$tmp_grep"
            echo "==================="
            ((fail++))
        fi
    fi
    echo
done

# Дополнительные тесты для проверки специфического поведения
echo "Testing specific edge cases..."
echo "=============================="

# Тест: несколько совпадений с контекстом (должны быть разделители --)
multi_test_cmd="./mygrep -A 1 test testcases/test1.txt"
grep_multi_test_cmd="grep -A 1 test testcases/test1.txt"

bash -c "$multi_test_cmd" > "$tmp_mygrep" 2>/dev/null
bash -c "$grep_multi_test_cmd" > "$tmp_grep" 2>/dev/null

if diff -u "$tmp_mygrep" "$tmp_grep" >/dev/null; then
    echo "✅  multi_match_context passed (separators match)"
    ((ok++))
else
    echo "❌  multi_match_context failed (separators differ)"
    echo "=== mygrep ==="
    cat "$tmp_mygrep"
    echo "=== grep ==="
    cat "$tmp_grep"
    ((fail++))
fi
echo

# Тест: проверка что разделители -- появляются только при контексте
no_context_cmd="./mygrep test testcases/test1.txt"
grep_no_context_cmd="grep test testcases/test1.txt"

bash -c "$no_context_cmd" > "$tmp_mygrep" 2>/dev/null
bash -c "$grep_no_context_cmd" > "$tmp_grep" 2>/dev/null

if diff -u "$tmp_mygrep" "$tmp_grep" >/dev/null; then
    echo "✅  no_context_separators passed (no separators when no context)"
    ((ok++))
else
    echo "❌  no_context_separators failed (unexpected separators)"
    ((fail++))
fi
echo

# Cleanup
rm -f "$tmp_mygrep" "$tmp_grep"
rm -f testcases/empty.txt

echo "================================="
echo "Tests completed:"
echo "Passed: $ok"
echo "Failed: $fail"
echo "================================="

# Выход с кодом 1, если есть неудачные тесты
(( fail == 0 )) || exit 1
