package main

import "strings"

// lineWriter adapts an io.Writer interface onto a per-write callback —
// used to route slog output straight into the GUI's log widget instead
// of stdout.
type lineWriter struct {
	onLine func(string)
}

func newLineWriter(onLine func(string)) *lineWriter {
	return &lineWriter{onLine: onLine}
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.onLine(string(p))
	return len(p), nil
}

func countLines(s string) int {
	if s == "" {
		return 0
	}

	return strings.Count(s, "\n") + 1
}

// trimToLastLines keeps only the last n lines of s.
func trimToLastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}

	return strings.Join(lines[len(lines)-n:], "\n")
}
