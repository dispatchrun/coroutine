package types

import (
	"encoding/binary"
	"math"
)

func (s *Serializer) appendVarint(v int) { s.buffer = binary.AppendVarint(s.buffer, int64(v)) }

func (d *Deserializer) varint() int {
	l, n := binary.Varint(d.buffer)
	if n <= 0 {
		panic("invalid varint")
	}
	d.buffer = d.buffer[n:]
	return int(l)
}

func (s *Serializer) appendBool(v bool) {
	c := uint8(0)
	if v {
		c = 1
	}
	s.appendUint8(c)
}

func (d *Deserializer) bool() bool {
	v := d.buffer[0] == 1
	d.buffer = d.buffer[1:]
	return v
}

func (s *Serializer) appendInt(v int)     { s.appendUint64(uint64(v)) }
func (s *Serializer) appendInt8(v int8)   { s.appendUint8(uint8(v)) }
func (s *Serializer) appendInt16(v int16) { s.appendUint16(uint16(v)) }
func (s *Serializer) appendInt32(v int32) { s.appendUint32(uint32(v)) }
func (s *Serializer) appendInt64(v int64) { s.appendUint64(uint64(v)) }

func (d *Deserializer) int() int     { return int(d.uint64()) }
func (d *Deserializer) int8() int8   { return int8(d.uint8()) }
func (d *Deserializer) int16() int16 { return int16(d.uint16()) }
func (d *Deserializer) int32() int32 { return int32(d.uint32()) }
func (d *Deserializer) int64() int64 { return int64(d.uint64()) }

func (s *Serializer) appendUint(v uint)       { s.appendUint64(uint64(v)) }
func (s *Serializer) appendUintptr(v uintptr) { s.appendUint64(uint64(v)) }

func (d *Deserializer) uint() uint       { return uint(d.uint64()) }
func (d *Deserializer) uintptr() uintptr { return uintptr(d.uint64()) }

func (s *Serializer) appendUint8(v uint8) {
	s.buffer = append(s.buffer, byte(v))
}

func (s *Serializer) appendUint16(v uint16) {
	s.buffer = binary.LittleEndian.AppendUint16(s.buffer, uint16(v))
}

func (s *Serializer) appendUint32(v uint32) {
	s.buffer = binary.LittleEndian.AppendUint32(s.buffer, uint32(v))
}

func (s *Serializer) appendUint64(v uint64) {
	s.buffer = binary.LittleEndian.AppendUint64(s.buffer, uint64(v))
}

func (d *Deserializer) uint8() uint8 {
	v := uint8(d.buffer[0])
	d.buffer = d.buffer[1:]
	return v
}

func (d *Deserializer) uint16() uint16 {
	v := binary.LittleEndian.Uint16(d.buffer[:2])
	d.buffer = d.buffer[2:]
	return v
}

func (d *Deserializer) uint32() uint32 {
	v := binary.LittleEndian.Uint32(d.buffer[:4])
	d.buffer = d.buffer[4:]
	return v
}

func (d *Deserializer) uint64() uint64 {
	v := binary.LittleEndian.Uint64(d.buffer[:8])
	d.buffer = d.buffer[8:]
	return v
}

func (s *Serializer) appendFloat32(v float32) { s.appendUint32(math.Float32bits(v)) }
func (s *Serializer) appendFloat64(v float64) { s.appendUint64(math.Float64bits(v)) }

func (d *Deserializer) float32() float32 { return math.Float32frombits(d.uint32()) }
func (d *Deserializer) float64() float64 { return math.Float64frombits(d.uint64()) }
