package serde_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stealthrocket/coroutine/serde"
)

type basicTest[T comparable] struct {
	name  string
	ser   func(s *serde.Serializer, x T, b []byte) []byte
	des   func(d *serde.Deserializer, b []byte) (T, []byte)
	cases []T
}

func (bt basicTest[T]) Run(t *testing.T) {
	t.Run(bt.name, func(t *testing.T) {
		for i, x := range bt.cases {
			x := x
			t.Run(fmt.Sprintf("%d %v", i, x), func(t *testing.T) {
				s := serde.EnsureSerializer(nil)
				d := serde.EnsureDeserializer(nil)
				var b []byte
				b = bt.ser(s, x, b)
				y, b := bt.des(d, b)

				if x != y {
					t.Fatalf("got '%v'; expected '%v'", y, x)
				}

				if len(b) != 0 {
					t.Fatalf("extra %d bytes: %v", len(b), b)
				}
			})
		}
	})
}

func TestBasicSerdes(t *testing.T) {
	basicTest[bool]{
		name: "bool",
		ser:  serde.SerializeBool,
		des:  serde.DeserializeBool,
		cases: []bool{
			true,
			false,
		},
	}.Run(t)

	basicTest[uint64]{
		name: "uint64",
		ser:  serde.SerializeUint64,
		des:  serde.DeserializeUint64,
		cases: []uint64{
			0,
			1,
			42,
			math.MaxUint64,
		},
	}.Run(t)

	basicTest[uint32]{
		name: "uint32",
		ser:  serde.SerializeUint32,
		des:  serde.DeserializeUint32,
		cases: []uint32{
			0,
			1,
			42,
			math.MaxUint32,
		},
	}.Run(t)

	basicTest[uint16]{
		name: "uint16",
		ser:  serde.SerializeUint16,
		des:  serde.DeserializeUint16,
		cases: []uint16{
			0,
			1,
			42,
			math.MaxUint16,
		},
	}.Run(t)

	basicTest[uint8]{
		name: "uint8",
		ser:  serde.SerializeUint8,
		des:  serde.DeserializeUint8,
		cases: []uint8{
			0,
			1,
			42,
			math.MaxUint8,
		},
	}.Run(t)

	basicTest[int64]{
		name: "int64",
		ser:  serde.SerializeInt64,
		des:  serde.DeserializeInt64,
		cases: []int64{
			0,
			1,
			-1,
			42,
			math.MinInt64,
			math.MaxInt64,
		},
	}.Run(t)

	basicTest[int32]{
		name: "int32",
		ser:  serde.SerializeInt32,
		des:  serde.DeserializeInt32,
		cases: []int32{
			0,
			1,
			-1,
			42,
			math.MinInt32,
			math.MaxInt32,
		},
	}.Run(t)

	basicTest[int16]{
		name: "int16",
		ser:  serde.SerializeInt16,
		des:  serde.DeserializeInt16,
		cases: []int16{
			0,
			1,
			-1,
			42,
			math.MinInt16,
			math.MaxInt16,
		},
	}.Run(t)

	basicTest[int8]{
		name: "int8",
		ser:  serde.SerializeInt8,
		des:  serde.DeserializeInt8,
		cases: []int8{
			0,
			1,
			-1,
			42,
			math.MinInt8,
			math.MaxInt8,
		},
	}.Run(t)

	basicTest[float32]{
		name: "float32",
		ser:  serde.SerializeFloat32,
		des:  serde.DeserializeFloat32,
		cases: []float32{
			0.0,
			1.0,
			-1.0,
			42.42,
			math.SmallestNonzeroFloat32,
			math.MaxFloat32,
		},
	}.Run(t)

	basicTest[float64]{
		name: "float64",
		ser:  serde.SerializeFloat64,
		des:  serde.DeserializeFloat64,
		cases: []float64{
			0.0,
			1.0,
			-1.0,
			42.42,
			math.SmallestNonzeroFloat64,
			math.MaxFloat64,
		},
	}.Run(t)

	basicTest[complex64]{
		name: "complex64",
		ser:  serde.SerializeComplex64,
		des:  serde.DeserializeComplex64,
		cases: []complex64{
			0i,
			1i,
			-1i,
			42i,
			(3 + 4i),
		},
	}.Run(t)

	basicTest[complex128]{
		name: "complex128",
		ser:  serde.SerializeComplex128,
		des:  serde.DeserializeComplex128,
		cases: []complex128{
			0i,
			1i,
			-1i,
			42i,
			(3 + 4i),
		},
	}.Run(t)

	basicTest[string]{
		name: "string",
		ser:  serde.SerializeString,
		des:  serde.DeserializeString,
		cases: []string{
			"",
			"hello",
			"world",
		},
	}.Run(t)
}
