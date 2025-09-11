#!/usr/bin/env bash
set -uo pipefail

MINISHELL=./shell
TMPDIR=$(mktemp -d)
REF_OUT=$TMPDIR/ref.out
REF_ERR=$TMPDIR/ref.err
REF_CODE=$TMPDIR/ref.code
MS_OUT=$TMPDIR/ms.out
MS_ERR=$TMPDIR/ms.err
MS_CODE=$TMPDIR/ms.code

# список команд для проверки
COMMANDS=(
  'echo Hello World'
  'echo $PATH'
  'pwd'
  'ls -la'
  'date'
  'whoami'
  'echo "Test line 2" > test.txt'
  'echo "Test line 3" >> test.txt'
  'echo "Test line 1" >> test.txt'
  'wc -l < test.txt'
  'cat < test.txt'
  'sort < test.txt > sorted.txt'
  'echo "apple\nbanana\ncherry\napple\nbanana" > fruits.txt'
  'echo "123\n456\n789\n123" > numbers.txt'
  'cat < fruits.txt'
  'cat < numbers.txt'
  'ls -la | grep txt'
  'echo "hello world" | wc -w'
  'cat fruits.txt | sort | uniq | wc -l'
  'ls -la | grep txt | wc -l'
  'ls -la | grep txt > txt_files.txt'
  'echo "success" && echo "also success"'
  'ls /nonexistent'
  'ls /nonexistent && echo "this won'\''t print"'
  'ls /nonexistent || echo "command failed"'
  'echo "first" && echo "second" || echo "won'\''t print"'
  'echo $NONEXISTENT_VARIABLE'
  'echo hello'
  'echo -n hello'
  'echo -nnnn hello'
  'echo -n'
  'echo'
)

# утилита для выполнения команды и сохранения результатов
run_cmd() {
  local shell_cmd="$1"
  local outfile="$2"
  local errfile="$3"
  local codefile="$4"
  local shell="$5"

  if [[ "$shell" == "bash" ]]; then
    bash -c "$shell_cmd" >"$outfile" 2>"$errfile"
    echo $? >"$codefile"
  else
    echo "$shell_cmd" | "$MINISHELL" >"$outfile" 2>"$errfile"
    echo $? >"$codefile"
  fi
}

# проверка результатов
compare_results() {
  local cmd="$1"

  if ! diff -u "$REF_OUT" "$MS_OUT"; then
    echo "❌ STDOUT differs for: $cmd"
  fi
  if ! diff -u "$REF_ERR" "$MS_ERR"; then
    echo "❌ STDERR differs for: $cmd"
  fi
}

# основной цикл
for cmd in "${COMMANDS[@]}"; do
  echo "▶️ Testing: $cmd"

  run_cmd "$cmd" "$REF_OUT" "$REF_ERR" "$REF_CODE" "bash"
  run_cmd "$cmd" "$MS_OUT" "$MS_ERR" "$MS_CODE" "minishell"

  compare_results "$cmd"
done

echo "✅ Tests finished. Temp dir: $TMPDIR"
