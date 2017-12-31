package az

import (
	"fmt"
	"io"
)

type Logger struct {
	writer io.Writer
}

func NewLogger(writer io.Writer) *Logger {
	return &Logger{
		writer: writer,
	}
}

func (l *Logger) Println(message string) {
	fmt.Fprintf(l.writer, "%s\n", message)
}
