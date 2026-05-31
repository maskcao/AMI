package opslog

import (
	"fmt"
	"sync"

	"github.com/ami/ami/service/internal/store"
)

var (
	mu sync.Mutex
	db *store.DB
)

// Init attaches persistent log storage (call once after store.Open).
func Init(d *store.DB) {
	mu.Lock()
	defer mu.Unlock()
	db = d
}

func record(level, source, message, detail string) {
	mu.Lock()
	d := db
	mu.Unlock()
	if d == nil {
		return
	}
	_ = d.InsertRuntimeLog(level, source, message, detail)
}

func Info(source, message string, detail ...string) {
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	record("info", source, message, d)
}

func Warn(source, message string, detail ...string) {
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	record("warn", source, message, d)
}

func Error(source, message string, detail ...string) {
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	record("error", source, message, d)
}

func Errorf(source, format string, args ...any) {
	Error(source, fmt.Sprintf(format, args...))
}
