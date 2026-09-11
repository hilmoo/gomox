// Package log builds a structured [slog.Logger] from a level and an output
// destination, both given as plain strings so they can come straight from
// config/env without any app-specific config type.
package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// New builds a *[slog.Logger] that writes JSON-formatted records at level to
// output.
//
// level is parsed case-insensitively via [slog.Level.UnmarshalText], so any
// of slog's standard names work: "debug", "info", "warn", "error" (as well
// as numeric offsets like "warn+2").
//
// output selects the destination:
//   - "" or "stdout" writes to [os.Stdout]
//   - "stderr" writes to [os.Stderr]
//   - any other value is treated as a file path and opened for appending,
//     creating it with mode 0o600 if it doesn't already exist
//
// New returns an error if level can't be parsed or, when output names a
// file, if that file can't be opened.
func New(level, output string) (*slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", level, err)
	}

	var w io.Writer
	switch strings.ToLower(output) {
	case "", "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return nil, fmt.Errorf("open log output %q: %w", output, err)
		}
		w = f
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: lvl,
	})

	return slog.New(handler), nil
}

// NewDefault behaves like [New], and additionally registers the resulting
// logger as the process-wide default via [slog.SetDefault]. Use it during
// startup so that package-level calls like slog.Info and any third-party
// code that logs through the default logger produce output consistent with
// the rest of the app.
func NewDefault(level, output string) (*slog.Logger, error) {
	logger, err := New(level, output)
	if err != nil {
		return nil, err
	}

	slog.SetDefault(logger)

	return logger, nil
}
