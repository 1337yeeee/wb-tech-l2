#!/usr/bin/env bash
set -Eeuo pipefail

# Сборка бинаря
go build -o sort sort.go

# Массив тестов: "<name>|<command>|<expected_file>"
TESTS=(
"t1_basic|./sort testcases/t1_basic.txt|testcases/expected/t1_basic.out"
"t2_numeric|./sort -n testcases/t2_numeric.txt|testcases/expected/t2_numeric.out"
"t3_reverse|./sort -r testcases/t3_reverse.txt|testcases/expected/t3_reverse.out"
"t4_unique|./sort -u testcases/t4_unique.txt|testcases/expected/t4_unique.out"
"t5_months|./sort -M testcases/t5_months.txt|testcases/expected/t5_months.out"
"t6_human|./sort -h testcases/t6_human.txt|testcases/expected/t6_human.out"
"t7_stdin|cat testcases/t1_basic.txt | ./sort|testcases/expected/t1_basic.out"
"t8_k2_tab|./sort -k 2 testcases/t8_k2_tab.txt|testcases/expected/t8_k2_tab.out"
"t9_nr_combo|./sort -n -r testcases/t2_numeric.txt|testcases/expected/t9_nr_combo.out"

# check flag tests
"t10_check_sorted|./sort -c testcases/expected/t1_basic.out|testcases/expected/empty.out"
"t11_check_unsorted|! ./sort -c testcases/t1_basic.txt|testcases/expected/empty.out"
"t12_check_numeric|./sort -c -n testcases/expected/t2_numeric.out|testcases/expected/empty.out"
"t13_check_numeric_unsorted|! ./sort -c -n testcases/t2_numeric.txt|testcases/expected/empty.out"
"t14_check_reverse|./sort -c -r testcases/expected/t3_reverse.out|testcases/expected/empty.out"
"t15_check_reverse_unsorted|! ./sort -c -r testcases/t3_reverse.txt|testcases/expected/empty.out"
"t16_check_months|./sort -c -M testcases/expected/t5_months.out|testcases/expected/empty.out"
"t17_check_months_unsorted|! ./sort -c -M testcases/t5_months.txt|testcases/expected/empty.out"
"t18_check_human|./sort -c -h testcases/expected/t6_human.out|testcases/expected/empty.out"
"t19_check_human_unsorted|! ./sort -c -h testcases/t6_human.txt|testcases/expected/empty.out"
)

ok=0
fail=0
tmp_out="$(mktemp)"

for t in "${TESTS[@]}"; do
    name="${t%%|*}"                  # имя теста
    rest="${t#*|}"                   # всё после первого |
    cmd="${rest%|*}"                 # команда (между первым и последним)
    expected="${rest##*|}"           # файл с ожидаемым результатом

    echo "▶️  $name: $cmd"

    if bash -lc "$cmd" > "$tmp_out"; then
        if diff -u "$tmp_out" "$expected"; then
            echo "✅  $name passed"
            ((ok++))
        else
            echo "❌  $name failed (output differs)"
            ((fail++))
        fi
    else
        echo "❌  $name failed (command error)"
        ((fail++))
    fi
    echo
done


rm -f "$tmp_out"

echo "===================="
echo "Passed: $ok"
echo "Failed: $fail"
echo "===================="

# Выход с кодом 1, если есть неудачные тесты
(( fail == 0 )) || exit 1
