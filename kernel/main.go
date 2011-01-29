package main

import "runtime"

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
//	runtime.EndCritical()
/*
	var initp Process
	initrd := make(Initrd)
	initrd["hello"] = testbinary[:]
	initrd["hello.txt"] = ([]byte)("Hello, World")[:]
	rootns := Namespace{NamespaceEntry{string: "/", Filesystem: initrd}}
	f, err := rootns.Open("/hello", ORD, 0)
	if err != nil {
		print("error opening /hello: ")
		println(err.String())
		for {}
	}
	err = initp.Exec(f)
	if err != nil {
		print("error executing /hello: ")
		println(err.String())
		for {}
	}
	initp.ns = rootns
	initp.ProcState.flags = 0x200
	initp.Run()
*/
	initpci()
	initbio()
	c := initide()
	b, err := BRead(&c.D[0], 0)
	if err != nil {
		println(err.String())
		for {}
	}
	b.Data[0] = 42
	err = b.Write()
	if err != nil {
		println(err.String())
		for {}
	}
}
