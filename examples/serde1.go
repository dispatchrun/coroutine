package examples

import (
	"fmt"

	"github.com/stealthrocket/ring/task"
)

type Struct1 struct {
	Str string
	Int int64
	X   interface{}
}

func create() {
	s := Struct1{
		X: true,
	}
	fmt.Println(s)
	s.X = Struct2{A: 42}
	fmt.Println(s)
	s.X = task.Config{}
}
