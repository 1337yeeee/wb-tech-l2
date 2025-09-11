package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type Command struct {
	name      string
	args      []string
	input     io.Reader
	output    io.Writer
	appendOut bool
	exitCode  int
	err       error
	shell     *Shell
	cmd       *exec.Cmd
}

// NewCommand создает экземпляр Command
func NewCommand(name string, args []string, input io.Reader, output io.Writer, appendOut bool, shell *Shell) *Command {
	return &Command{
		name:      name,
		args:      args,
		input:     input,
		output:    output,
		appendOut: appendOut,
		shell:     shell,
	}
}

// ExecuteCommand выполняет команду будь она builtin или внешней
func (c *Command) ExecuteCommand() {
	if len(c.args) == 0 {
		c.exitCode = 1
		c.err = errors.New("empty command")
		return
	}

	// builtin
	if ok, err := c.handleBuiltin(); ok {
		if err != nil {
			c.exitCode = 1
			c.err = err
		} else {
			c.exitCode = 0
			c.err = nil
		}
		return
	}

	// external
	err := c.handleExternal()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			c.exitCode = exitErr.ExitCode()
		} else {
			c.exitCode = 1
		}
		c.err = err
	} else {
		c.exitCode = 0
		c.err = nil
	}
}

// handleBuiltin выполняет builtin команду
func (c *Command) handleBuiltin() (bool, error) {
	name := c.args[0]
	args := c.args[1:]

	switch name {
	case "cd":
		if len(args) < 1 {
			home := os.Getenv("HOME")
			if home == "" {
				home = "/"
			}
			return true, os.Chdir(home)
		}
		return true, os.Chdir(args[0])
	case "pwd":
		wd, err := os.Getwd()
		fmt.Fprintln(c.output, wd)
		return true, err
	case "echo":
		return c.echo(args)
	case "kill":
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return true, err
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return true, err
		}
		return true, process.Kill()
	case "ps":
		fmt.Fprintf(c.output, "%6s %6s %8s %s\n", "PID", "PPID", "OS", "ARCH")
		fmt.Fprintf(c.output, "%6v %6v %8s %s\n", os.Getpid(), os.Getppid(), runtime.GOOS, runtime.GOARCH)
		return true, nil
	case "exit":
		c.shell.Stop()
		return true, nil
	}
	return false, nil
}

// handleExternal выполняет внешние команды через exec.Command.Run()
func (c *Command) handleExternal() error {
	cmd := exec.Command(c.args[0], c.args[1:]...)
	cmd.Stdin = c.input
	cmd.Stdout = c.output
	cmd.Stderr = os.Stderr

	// чтобы ловить Ctrl+C
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	c.cmd = cmd
	c.shell.appendCmd(cmd)

	return cmd.Start()
}

func (c *Command) echo(args []string) (bool, error) {
	noNewline := false
	argsStart := 0

	// проверяем флаг -n (может быть несколько раз)
	for argsStart < len(args) {
		if strings.HasPrefix(args[argsStart], "-n") && allN(args[argsStart]) {
			noNewline = true
			argsStart++
		} else {
			break
		}
	}

	text := strings.Join(args[argsStart:], " ")
	text = strings.ReplaceAll(text, "\\n", "\n")

	if noNewline {
		fmt.Fprint(c.output, text)
	} else {
		fmt.Fprintln(c.output, text)
	}
	return true, nil
}

func allN(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] != 'n' {
			return false
		}
	}
	return true
}
