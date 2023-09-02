package coroutine

import (
	"encoding/binary"
	"fmt"
)

// This file defines Serializable versions of builtin types.

// Int is a Serializable int.
type Int int

var _ Serializable = Int(0)
var _ Deserializable = (*Int)(nil)
var _ UnmarshalSerializable = UnmarshalInt

func (i Int) MarshalAppend(b []byte) ([]byte, error) {
	return binary.AppendVarint(b, int64(i)), nil
}

func (i *Int) Unmarshal(b []byte) (int, error) {
	value, n := binary.Varint(b)
	if n <= 0 || int64(Int(value)) != value {
		return 0, fmt.Errorf("invalid Int: %v", b)
	}
	*i = Int(value)
	return n, nil
}

func UnmarshalInt(b []byte) (_ Serializable, n int, err error) {
	var value Int
	n, err = value.Unmarshal(b)
	return value, n, err
}

func init() {
	RegisterSerializableConstructor(Int(0), UnmarshalInt)
}
