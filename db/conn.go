package db

import (
	"bytes"
	"io"
)

// Conn is a single connection to Addr
type Conn struct {
	Addr string
}

// Close pretends to Close but does nothing
func (c Conn) Close() error {
	return nil
}

// Open pretends to Open but just does nothing
func (c Conn) Open() error {
	return nil
}

// SendAndReceive pretends to SendAndReceive but returns a reader for the input bytes + 👋
func (c Conn) SendAndReceive(in []byte) (io.Reader, error) {
	b := new(bytes.Buffer)
	res := append(in, []byte("👋")...)
	if _, err := b.Write(res); err != nil {
		return nil, err
	}
	return b, nil
}
