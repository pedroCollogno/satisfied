// A custom colored [slog.Handler] inspired by [tint package](https://github.com/lmittmann/tint)
// but simplified.

package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
)

// ANSI modes
const (
	ansiReset          = "\033[0m"
	ansiFaint          = "\033[2m"
	ansiResetFaint     = "\033[22m"
	ansiBrightRed      = "\033[91m"
	ansiBrightGreen    = "\033[92m"
	ansiBrightYellow   = "\033[93m"
	ansiBrightBlue     = "\033[94m"
	ansiBrightRedFaint = "\033[91;2m"
)

const (
	lowerLvlStr  = "TRC-"
	traceLvlStr  = "TRC"
	debugLvlStr  = ansiBrightBlue + "DBG" + ansiReset
	infoLvlStr   = ansiBrightGreen + "INF" + ansiReset
	warnLvlStr   = ansiBrightYellow + "WRN" + ansiReset
	errorLvlStr  = ansiBrightRed + "ERR" + ansiReset
	fatalLvlStr  = ansiBrightRed + "FTL" + ansiReset
	higherLvlStr = "FTL+"
)

type Handler struct {
	w     io.Writer
	level slog.Level
}

func NewHandler(w io.Writer, level slog.Level) *Handler {
	return &Handler{w: w, level: level}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	src := formatSource(r.PC)
	lvl := h.formatLevel(r.Level)
	s := fmt.Sprintf("%-30s %s %-40s", src, lvl, r.Message)
	r.Attrs(func(attr slog.Attr) bool {
		attr.Value = attr.Value.Resolve()
		if attr.Value.Kind() == slog.KindString {
			attr.Value = slog.StringValue(strconv.Quote(attr.Value.String()))
		}
		s = fmt.Sprintf("%s %s%s=%s%s", s, ansiFaint, attr.Key, ansiReset, attr.Value)
		return true
	})
	_, err := fmt.Fprintln(h.w, s)
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("not implemented")
}

func (h *Handler) WithGroup(name string) slog.Handler {
	panic("not implemented")
}

func (h Handler) formatLevel(lvl slog.Level) string {
	switch lvl {
	case TraceLevel:
		return traceLvlStr
	case DebugLevel:
		return debugLvlStr
	case InfoLevel:
		return infoLvlStr
	case WarnLevel:
		return warnLvlStr
	case ErrorLevel:
		return errorLvlStr
	case FatalLevel:
		return fatalLvlStr
	default:
		if lvl < TraceLevel {
			return lowerLvlStr
		} else {
			return higherLvlStr
		}
	}
}

func formatSource(pc uintptr) string {
	fs := runtime.CallersFrames([]uintptr{pc})
	f, more := fs.Next()
	if more {
		f, _ = fs.Next()
	}
	if f.File != "" {
		// find the idx of the last directory start
		idx := 0
		firstFound := false
		for i := len(f.File) - 1; i >= 0; i-- {
			if f.File[i] == '/' {
				if !firstFound {
					firstFound = true
				} else {
					idx = i + 1
					break
				}
			}
		}
		return fmt.Sprintf("%s%s:%d%s", ansiFaint, f.File[idx:], f.Line, ansiReset)
	}
	return ""
}
