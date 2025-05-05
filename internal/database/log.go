package database

import (
	"bytes"
	"encoding/gob"
)

type WALRecord struct {
	Command   Command
	Arguments []string
}

func (wr *WALRecord) Encode(buffer *bytes.Buffer) error {
	encoder := gob.NewEncoder(buffer)
	return encoder.Encode(*wr)
}

func (wr *WALRecord) Decode(buffer *bytes.Buffer) error {
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(wr)
}
