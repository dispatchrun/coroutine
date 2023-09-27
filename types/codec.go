package types

import (
	"fmt"
	"time"
)

func init() {
	Register[time.Time](serializeTime, deserializeTime)
}

func serializeTime(s *Serializer, x *time.Time) error {
	data, err := x.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal time.Time: %w", err)
	}

	SerializeT(s, data)
	return nil
}

func deserializeTime(d *Deserializer, x *time.Time) error {
	var b []byte
	DeserializeTo(d, &b)
	return x.UnmarshalBinary(b)
}
