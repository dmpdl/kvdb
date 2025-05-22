package replication

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Request struct {
	PreviousSegment string
}

type Response struct {
	Segment string
	Data    []byte
	Error   string
}

func Encode[ProtocolObject Request | Response](object *ProtocolObject) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(object); err != nil {
		return []byte(fmt.Sprintf("failed to encode: %s", err.Error()))
	}

	return buffer.Bytes()
}

func Decode[ProtocolObject Request | Response](object *ProtocolObject, data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(&object); err != nil {
		return fmt.Errorf("failed to decode object: %w", err)
	}

	return nil
}

func ErrResponse(err error) []byte {
	replicationResponse := Response{
		Error: err.Error(),
	}

	return Encode(&replicationResponse)
}
