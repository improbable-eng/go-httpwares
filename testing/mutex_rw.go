package httpwares_testing

import (
	"io"
	"sync"
)

// MutexReadWriter is a io.ReadWriter that can be read and worked on from multiple go routines.
type MutexReadWriter struct {
	sync.Mutex
	rw io.ReadWriter
}

// NewMutexReadWriter creates a new thread-safe io.ReadWriter.
func NewMutexReadWriter(rw io.ReadWriter) *MutexReadWriter {
	return &MutexReadWriter{rw: rw}
}

// Write implements the io.Writer interface.
func (gb *MutexReadWriter) Write(p []byte) (int, error) {
	gb.Lock()
	defer gb.Unlock()
	return gb.rw.Write(p)
}

// Read implements the io.Reader interface.
func (gb *MutexReadWriter) Read(p []byte) (int, error) {
	gb.Lock()
	defer gb.Unlock()
	return gb.rw.Read(p)
}
