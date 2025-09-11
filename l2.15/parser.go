package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParsedCommand представляет распарсенную команду
type ParsedCommand struct {
	Name      string
	Args      []string
	Input     io.Reader
	Output    io.Writer
	AppendOut bool
}

// ToCommand создает Command из ParsedCommand
func (c *ParsedCommand) ToCommand(shell *Shell) *Command {
	return NewCommand(
		c.Name,
		c.Args,
		c.Input,
		c.Output,
		c.AppendOut,
		shell,
	)
}

// Pipeline представляет последовательность команд, соединенных пайпами
type Pipeline struct {
	Commands []ParsedCommand
}

// LogicalCommand представляет логическую последовательность команд
type LogicalCommand struct {
	Pipelines []Pipeline
	Operators []string // "&&" или "||"
}

// ParseLine парсит всю строку команды
func ParseLine(line string) (*LogicalCommand, error) {
	if line == "" {
		return nil, errors.New("empty command")
	}

	parts, ops := splitByLogicalOps(line)

	logicalCommand := &LogicalCommand{Operators: ops}

	for _, part := range parts {
		pipeline, err := parsePipeline(part)
		if err != nil {
			return nil, fmt.Errorf("parse pipeline error: %w", err)
		}
		logicalCommand.Pipelines = append(logicalCommand.Pipelines, *pipeline)
	}

	return logicalCommand, nil
}

// parsePipeline парсит пайплайн команд
func parsePipeline(pipeSte string) (*Pipeline, error) {
	cmdStrs := splitByPipe(pipeSte)
	if len(cmdStrs) == 0 {
		return nil, errors.New("empty pipeline")
	}

	pipeline := &Pipeline{}

	for _, cmdStr := range cmdStrs {
		cmd, err := parseSingleCommand(cmdStr)
		if err != nil {
			return nil, fmt.Errorf("parse command '%s': %w", cmdStr, err)
		}
		pipeline.Commands = append(pipeline.Commands, *cmd)
	}

	return pipeline, nil
}

// parseSingleCommand парсит одну команду с аргументами и редиректами
func parseSingleCommand(cmdStr string) (*ParsedCommand, error) {
	tokens, err := splitArgs(cmdStr)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("empty command")
	}

	var input io.Reader
	var output io.Writer
	var appendOut bool
	var inFile, outFile string
	var args []string

	// Проходим по токенам и ищем редиректы
	i := 0
	for i < len(tokens) {
		t := tokens[i]

		switch t {
		case "<":
			if i+1 >= len(tokens) {
				return nil, errors.New("syntax error: < without file")
			}
			inFile = tokens[i+1]
			i += 2
		case ">":
			if i+1 >= len(tokens) {
				return nil, errors.New("syntax error: > without file")
			}
			outFile = tokens[i+1]
			appendOut = false
			i += 2
		case ">>":
			if i+1 >= len(tokens) {
				return nil, errors.New("syntax error: >> without file")
			}
			outFile = tokens[i+1]
			appendOut = true
			i += 2
		default:
			args = append(args, t)
			i++
		}
	}

	if len(args) == 0 {
		return nil, errors.New("syntax error: command not found")
	}

	if inFile != "" {
		input, err = os.Open(inFile)
		if err != nil {
			return nil, fmt.Errorf("open input file '%s': %w", inFile, err)
		}
	}

	if outFile != "" {
		var err error
		if appendOut {
			output, err = os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			output, err = os.Create(outFile)
		}
		if err != nil {
			return nil, fmt.Errorf("open output file '%s': %w", outFile, err)
		}
	}

	return &ParsedCommand{
		Name:      args[0],
		Args:      args,
		Input:     input,
		Output:    output,
		AppendOut: appendOut,
	}, nil
}

// splitByLogicalOps разделяет строку по && и ||
func splitByLogicalOps(line string) ([]string, []string) {
	var parts []string
	var ops []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	runes := []rune(line)
	i := 0

	for i < len(runes) {
		r := runes[i]

		// обработка кавычек
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			cur.WriteRune(r)
			i++
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			cur.WriteRune(r)
			i++
			continue
		}

		// обработка операторов вне кавычек
		if !inSingle && !inDouble && i+1 < len(runes) {
			if r == '|' && runes[i+1] == '|' {
				parts = append(parts, strings.TrimSpace(cur.String()))
				cur.Reset()
				ops = append(ops, "||")
				i += 2
				continue
			} else if r == '&' && runes[i+1] == '&' {
				parts = append(parts, strings.TrimSpace(cur.String()))
				cur.Reset()
				ops = append(ops, "&&")
				i += 2
				continue
			}
		}

		cur.WriteRune(r)
		i++
	}

	if cur.Len() > 0 {
		parts = append(parts, strings.TrimSpace(cur.String()))
	}

	return parts, ops
}

// splitByPipe разделяет строку по |
func splitByPipe(line string) []string {
	var parts []string
	var cur strings.Builder
	inSingle, inDouble := false, false

	for _, r := range line {
		// обработка кавычек
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			cur.WriteRune(r)
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			cur.WriteRune(r)
			continue
		}

		// разделение по пайпу вне кавычек
		if !inSingle && !inDouble && r == '|' {
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}

		cur.WriteRune(r)
	}

	if cur.Len() > 0 {
		parts = append(parts, strings.TrimSpace(cur.String()))
	}

	return parts
}

// splitArgs разделяет строку на аргументы, учитывая кавычки и экранирование
func splitArgs(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	escaped := false

	for _, r := range s {
		x := string(r)
		if x == "˜" {
			return nil, errors.New("syntax error: invalid argument")
		}
		switch {
		case escaped:
			cur.WriteRune('\\')
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			// кавычки не включаем
		case r == '"' && !inSingle:
			inDouble = !inDouble
			// кавычки не включаем
		case r == ' ' || r == '\t':
			if inSingle || inDouble {
				cur.WriteRune(r)
			} else {
				if cur.Len() > 0 {
					args = append(args, cur.String())
					cur.Reset()
				}
			}
		default:
			cur.WriteRune(r)
		}
	}

	if escaped {
		return nil, errors.New("unfinished escape sequence")
	}
	if inSingle {
		return nil, errors.New("unterminated single quote")
	}
	if inDouble {
		return nil, errors.New("unterminated double quote")
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}

	return args, nil
}
