package adapter

import (
	"encoding/json"
	"os"
	"sync"

	"todoe/internal/audit/domain"
)

// FileWriter writes audit entries to a JSON-lines file as a fallback
// when MongoDB is unavailable.
type FileWriter struct {
	mu   sync.Mutex
	path string
}

func NewFileWriter(path string) *FileWriter {
	return &FileWriter{path: path}
}

func (w *FileWriter) Write(entry domain.AuditEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = f.Write(data)
	return err
}
