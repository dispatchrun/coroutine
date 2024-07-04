package types

func (s *Serializer) appendBool(v bool) {
	c := byte(0)
	if v {
		c = 1
	}
	s.buffer = append(s.buffer, c)
}

func (d *Deserializer) bool() bool {
	v := d.b[0] == 1
	d.b = d.b[1:]
	return v
}
