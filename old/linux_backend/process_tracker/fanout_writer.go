package process_tracker

import (
	"errors"
	"io"
	"sync"
)

type fanoutWriter struct {
	sinks  []io.Writer
	closed bool //TODO: Is this needed?
	sinksL sync.Mutex
}

func (w *fanoutWriter) Write(data []byte) (int, error) {
	w.sinksL.Lock()
	defer w.sinksL.Unlock()

	if w.closed {
		return 0, errors.New("write after close")
	}

	// the sinks should be nonblocking and never actually error;
	// we can assume lossiness here, and do this all within the lock
	for _, s := range w.sinks {
		s.Write(data)
	}

	return len(data), nil
}

func (w *fanoutWriter) AddSink(sink io.Writer) {
	w.sinksL.Lock()
	defer w.sinksL.Unlock()

	if !w.closed {
		w.sinks = append(w.sinks, sink)
	}
}
