package log

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	AllLevel slog.Level = iota
	TraceLevel
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var level slog.Level

// Initializes logging
func Init(lvl slog.Level, colored bool) {
	level = lvl
	slog.SetLogLoggerLevel(lvl)
	rl.SetTraceLogLevel(rl.TraceLogLevel(lvl))
	rl.SetTraceLogCallback(func(l int, message string) { Log(slog.Level(l), message, "source", "raylib") })
	var handler slog.Handler
	if colored {
		handler = NewHandler(os.Stderr, lvl)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, nil)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// WillTrace returns true if [TraceLevel] logs will be written
func WillTrace() bool { return level <= TraceLevel }

// Log at [TraceLevel]
func Trace(msg string, args ...any) { log(TraceLevel, msg, args...) }

// Log at [DebugLevel]
func Debug(msg string, args ...any) { log(DebugLevel, msg, args...) }

// Log at [InfoLevel]
func Info(msg string, args ...any) { log(InfoLevel, msg, args...) }

// Log at [WarnLevel]
func Warn(msg string, args ...any) { log(WarnLevel, msg, args...) }

// Log at [ErrorLevel]
func Error(msg string, args ...any) { log(ErrorLevel, msg, args...) }

// Log at [FatalLevel]
func Fatal(msg string, args ...any) { log(FatalLevel, msg, args...) }

// Log at given [slog.Level]
func Log(level slog.Level, msg string, args ...any) {
	log(level, msg, args...)
}

// copied and adapted from [log/slog] package
// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func log(lvl slog.Level, msg string, args ...any) {
	l := slog.Default()
	ctx := context.Background()

	if !l.Enabled(ctx, lvl) {
		return
	}
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc := pcs[0]
	r := slog.NewRecord(time.Now(), lvl, msg, pc)
	r.Add(args...)
	_ = l.Handler().Handle(ctx, r)
}
