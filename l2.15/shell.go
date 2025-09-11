package main

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Shell мини интерпретатор shell команд
type Shell struct {
	running      bool
	currentCmds  []*exec.Cmd // Текущие запущенные процессы (для пересылки сигналов)
	mu           sync.Mutex
	sigintChan   chan os.Signal
	lastExitCode int
}

// NewShell создает экземпляр Shell
func NewShell() *Shell {
	sh := &Shell{
		running:    false,
		sigintChan: make(chan os.Signal, 1),
	}
	signal.Notify(sh.sigintChan, os.Interrupt)
	go sh.handleSignals()
	return sh
}

// handleSignals пересылает SIGINT запущенным процессам, не завершая shell
func (s *Shell) handleSignals() {
	for range s.sigintChan {
		s.mu.Lock()
		cmds := s.currentCmds
		s.currentCmds = nil
		s.mu.Unlock()

		// если есть запущенные команды — отправляем SIGINT их process group
		if len(cmds) > 0 {
			for _, c := range cmds {
				if c != nil && c.Process != nil {
					pgid, err := syscall.Getpgid(c.Process.Pid)
					if err == nil {
						// отправляем сигнал всей группе процессов
						_ = syscall.Kill(-pgid, syscall.SIGINT)
					} else {
						_ = c.Process.Signal(os.Interrupt)
					}
				}
			}
		}
		// очищаем список текущих команд
		s.currentCmds = nil
		s.hello()
	}
}

// Run запускает shell: читает Stdin и выполняет команды
func (s *Shell) Run() {
	s.running = true

	scanner := bufio.NewScanner(os.Stdin)

	for s.running {
		// выводим приглашение командной строки
		s.hello()

		// читаем команду
		if !scanner.Scan() {
			// EOF (Ctrl+D / cmd + D)
			break
		}

		// чистим входную строку от лидирующий и следующих пробелов
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// подставляем переменные окружения
		line = expandEnvVars(line)

		// обрабатываем команду
		s.lastExitCode = 0
		s.executeLine(line)
		os.Exit(s.lastExitCode)

		// очищаем список текущих команд после выполнения
		s.currentCmds = nil
	}
}

// Stop останавливает shell интерпретатор
func (s *Shell) Stop() {
	s.running = false
}

// hello выводит приглашение командной строки
func (s *Shell) hello() {
	if term.IsTerminal(syscall.Stdin) {
		wd, _ := os.Getwd()
		fmt.Printf("\n[%s]$ ", filepath.Base(wd))
	}
}

// executeLine обрабатывает строку как команду
func (s *Shell) executeLine(line string) {
	logicalCommand, err := ParseLine(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		s.lastExitCode = 1
		return
	}

	exitCode, err := s.executeLogicalCommand(logicalCommand)

	s.lastExitCode = exitCode
}

// executeLogicalCommand обрабатывает LogicalCommand
func (s *Shell) executeLogicalCommand(logicalCmd *LogicalCommand) (int, error) {
	var exitCode int
	var err error

	i := 0
	for i < len(logicalCmd.Pipelines) {
		// выполнение пайплайна
		exitCode, err = s.executePipeline(&logicalCmd.Pipelines[i])

		// выполнена последняя команда
		if i == len(logicalCmd.Pipelines)-1 {
			break
		}

		// команда завершилась с ошибкой и текущий логический оператор &&:
		// пропускаем все следующие части, которые соединяет оператор &&
		if s.lastExitCode != 0 && i < len(logicalCmd.Operators) && logicalCmd.Operators[i] == "&&" {
			for ; i < len(logicalCmd.Operators) && logicalCmd.Operators[i] == "&&"; i++ {
			}
			// После цикла i либо i перешел границу количества операндов,
			// либо i является индексом оператора || и операнда перед ||
			// этот операнд не выполняется, так как является частью цепочки И,
			// в которой был возвращен ненулевой exitCode. Поэтому пропускаем его
			i++
			continue
		}

		// команда завершилась с ошибкой и текущий оператор ||
		// идем дальше
		if s.lastExitCode != 0 && logicalCmd.Operators[i] == "||" {
			i++
			continue
		}

		// команда завершилась без ошибки и текущий оператор &&
		// идем дальше, проверяя все операнды, связанные операцией &&
		if s.lastExitCode == 0 && logicalCmd.Operators[i] == "&&" {
			i++
			continue
		}

		// команда завершилась без ошибки и текущий оператор не &&
		// завершаем выполнение составной команды - успех
		break
	}

	return exitCode, err
}

func (s *Shell) executePipeline(pipeline *Pipeline) (int, error) {
	if len(pipeline.Commands) == 0 {
		return 0, nil
	}

	var prevReader io.Reader
	var commands []*Command
	var wg sync.WaitGroup

	// Создаем все команды
	for i := range pipeline.Commands {
		commands = append(commands, pipeline.Commands[i].ToCommand(s))
	}

	// Выполняем конвейер
	for i := range pipeline.Commands {
		cmdLocal := commands[i]

		// stdin
		switch {
		case cmdLocal.input != nil:
			// уже задан редиректом <
		case prevReader != nil:
			cmdLocal.input = prevReader // вход от предыдущего пайпа
		default:
			cmdLocal.input = os.Stdin // стандартный вход
		}

		// stdout
		var pipeWriter io.WriteCloser
		if i == len(commands)-1 {
			// последняя команда
			if cmdLocal.output == nil {
				cmdLocal.output = os.Stdout
			}
			// цепочку дальше не строим
			prevReader = nil
		} else {
			// не последняя команда
			if cmdLocal.output == nil {
				// создаём пайп только если нет редиректа > или >>
				pr, pw := io.Pipe()
				pipeWriter = pw
				cmdLocal.output = pw
				prevReader = pr
			} else {
				// у команды вывод в файл — дальше по пайпу ничего не пойдёт
				prevReader = nil
			}
		}

		wg.Add(1)
		go func(c *Command, w io.WriteCloser, r io.Reader) {
			defer wg.Done()
			// закрываем writer, чтобы downstream получил EOF
			defer func() {
				if w != nil && w != os.Stdout {
					_ = w.Close()
				}
				if r != os.Stdin {
					if rc, ok := r.(io.ReadCloser); ok {
						_ = rc.Close() // вот тут закрываем reader
					}
				}
			}()

			// запуск команды
			c.ExecuteCommand()

			// если это внешняя команда — дождаться
			if c.cmd != nil {
				if err := c.cmd.Wait(); err != nil {
					c.err = err
					if ee, ok := err.(*exec.ExitError); ok {
						c.exitCode = ee.ExitCode()
					} else {
						c.exitCode = 1
					}
				}
			}
		}(cmdLocal, pipeWriter, cmdLocal.input)
	}

	// дождаться всех
	wg.Wait()

	// статус и ошибка последней команды
	last := commands[len(commands)-1]
	s.lastExitCode = last.exitCode
	return last.exitCode, last.err
}

func (s *Shell) appendCmd(cmd *exec.Cmd) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentCmds = append(s.currentCmds, cmd)
}

// expandEnvVars — заменяет $VAR и ${VAR}
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		if key == "" {
			return "$"
		}
		return os.Getenv(key)
	})
}
