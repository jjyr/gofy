package runtime

import "unsafe"

func dummy() {
	var used unsafe.Pointer
	_ = used
}

func throwinit() {
	for {
	}
}

type itab struct {
	Itype  *Type
	Type   *Type
	link   *itab
	bad    int32
	unused int32
	Fn     func() // TODO: [0]func()
}
