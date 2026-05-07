package adapter

import (
	"time"

	"todoe/internal/sla/domain"
)

// Stopwatch tracks the duration of an event and writes to CSV on stop.
type Stopwatch struct {
	writer  *CSVWriter
	id      string
	typ     string
	started time.Time
}

// NewStopwatch creates a running stopwatch for the given event ID and type.
func NewStopwatch(writer *CSVWriter, id, typ string) *Stopwatch {
	return &Stopwatch{writer: writer, id: id, typ: typ, started: time.Now()}
}

// Stop records the elapsed time and returns the SLAEntry.
func (s *Stopwatch) Stop() domain.SLAEntry {
	entry := domain.SLAEntry{
		ID:        s.id,
		EventType: s.typ,
		Duration:  time.Since(s.started),
	}
	s.writer.Write(entry)
	return entry
}
