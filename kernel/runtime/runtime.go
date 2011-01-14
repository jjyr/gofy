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
