package serde

import (
	"fmt"
	"time"
)

func init() {
	RegisterTypeWithSerde[time.Time](serializeTime, deserializeTime)
}

func serializeTime(s *Serializer, x *time.Time) error {
	// see: https://github.com/golang/go/blob/38b2c06e144c6ea7087c575c76c66e41265ae0b7/src/time/time.go#L1240
	off := 1
	_, offset := x.Zone()
	if offset%60 != 0 {
		off = 0
	}

	b := make([]byte, 16)

	data, err := x.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal time.Time: %w", err)
	}
	copy(b[off:], data)

	Serialize(s, b)
	return nil
}

func deserializeTime(d *Deserializer, x *time.Time) error {
	buf := make([]byte, 16)
	DeserializeTo(d, &buf)
	off := 1
	if buf[0] > 0 {
		off = 0
	}
	return x.UnmarshalBinary(buf[off:])
}
