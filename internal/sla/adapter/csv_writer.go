package adapter

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"

	"todoe/internal/sla/domain"
)

type CSVWriter struct {
	mu     sync.Mutex
	file   *os.File
	writer *csv.Writer
}

func NewCSVWriter(path string) (*CSVWriter, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	w := &CSVWriter{file: f, writer: csv.NewWriter(f)}

	// Write header if file is empty
	if info, err := f.Stat(); err == nil && info.Size() == 0 {
		w.writer.Write([]string{"id", "event_type", "duration_ms"})
		w.writer.Flush()
	}

	return w, nil
}

func (w *CSVWriter) Write(entry domain.SLAEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.writer.Write([]string{
		entry.ID,
		entry.EventType,
		strconv.FormatInt(entry.Duration.Milliseconds(), 10),
	})
	w.writer.Flush()
	return err
}

func (w *CSVWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Flush()
}

func (w *CSVWriter) Close() error {
	w.Flush()
	return w.file.Close()
}
