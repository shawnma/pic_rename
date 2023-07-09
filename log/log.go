package log

import (
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type coloredPrint struct {
	color  *color.Color
	output io.Writer
}

func (c *coloredPrint) Log(a string, args ...any) {
	if !strings.HasSuffix(a, "\n") {
		a = a + "\n"
	}
	c.color.Fprintf(c.output, a, args...)
}

type Log struct {
	warning coloredPrint
	err     coloredPrint
	info    coloredPrint
	debug   coloredPrint
}

func New() *Log {
	return &Log{
		warning: coloredPrint{color.New(color.FgYellow), os.Stderr},
		err:     coloredPrint{color.New(color.FgHiRed), os.Stderr},
		info:    coloredPrint{color.New(color.FgCyan), os.Stderr},
		debug:   coloredPrint{color.New(), os.Stdout},
	}
}

func (l *Log) Info(a string, args ...any) {
	l.info.Log(a, args...)
}

func (l *Log) Warn(a string, args ...any) {
	l.warning.Log(a, args...)
}

func (l *Log) Error(a string, args ...any) {
	l.err.Log(a, args...)
}

func (l *Log) Debug(a string, args ...any) {
	l.debug.Log(a, args...)
}
