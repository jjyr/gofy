package main

import "runtime"

var testbinary [45]byte = [45]byte {
	0x00, 0x00, 0x8a, 0x97, 0x00, 0x00, 0x00, 0x05,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xe9, 0xfb, 0xff, 0xff, 0xff, 
}

func hex(n uint64, b bool) string {
	if n == 0 {
		if b {
			return "0"
		} else {
			return ""
		}
	}
	return hex(n / 16, false) + string("0123456789ABCDEF"[n & 15])
}

func fuck(s string) {
	println("SHIT IS BROKEN")
	println(s)
	runtime.Halt()
}


func main() {
//	var initp Process
	for {}
	initrd := make(Initrd)
	initrd["hello"] = testbinary[:]
	rootns := Namespace{NamespaceEntry{string: "/", Filesystem: initrd}}
	_, err := rootns.Open("/hello", ORD, 0)
	if err != nil {
		println(err.String())
		for {}
	}
/*
	err = initp.Exec(f)
	if err != nil {
		println(err.String())
		for {}
	}
	initp.Run()*/
}
