package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type HistoryEntry struct {
	Timestamp string
	Hostname  string
	Action    string
	Params    string
	Result    string
}

type ActionLogger struct {
	mu       sync.Mutex
	filePath string
}

var actionLogger = NewActionLogger()

func NewActionLogger() *ActionLogger {
	homedir, _ := os.UserHomeDir()
	return &ActionLogger{
		filePath: filepath.Join(homedir, ".lazy-hole", "history.log"),
	}
}

func (l *ActionLogger) Log(hostname, action, params, result string) {
	l.mu.Lock()
	defer l.mu.Unlock() // unlock after function done

	// reate new dir if not exists
	dir := filepath.Dir(l.filePath)
	os.MkdirAll(dir, 0755)

	// Open file in append mode
	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return // silent fail - logging should not crash app xD even it fails
	}

	defer f.Close() // close file after func done

	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s | %s | %s | %s | %s\n", timestamp, hostname, action, params, result)
	f.WriteString(line)
}

func (l *ActionLogger) ReadHistory() ([]HistoryEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no file, no history
		}
		return nil, err
	}

	var entries []HistoryEntry
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " | ", 5)
		if len(parts) != 5 {
			continue // skip line err!
		}
		entries = append(entries, HistoryEntry{
			Timestamp: parts[0],
			Hostname:  parts[1],
			Action:    parts[2],
			Params:    parts[3],
			Result:    parts[4],
		})
	}
	return entries, nil
}
