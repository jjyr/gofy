package main

import (
	"os"
	"fmt"
	"encoding/binary"
)

type Exec struct {
	Magic, Text, Data, Bss, Syms, Entry, Spsz, Pcsz uint32
}

func main() {
	if len(os.Args) != 2 {
		println("wrong number of arguments")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	var h Exec
	err = binary.Read(f, binary.BigEndian, &h)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	if h.Magic != 0x8a97 {
		println("wrong magic")
		os.Exit(1)
	}
	println(h.Syms)
	f.Seek(int64(40 + h.Text + h.Data), 0)
	for i := 0; i < int(h.Syms); {
		var value uint32
		var t [1]byte
		var name [100]byte
		err = binary.Read(f, binary.BigEndian, &value)
		if err != nil {
			println(err)
			os.Exit(1)
		}
		_, err = f.Read(t[:])
		if err != nil {
			println(err)
			os.Exit(1)
		}
		n := 0
		for {
			_, err = f.Read(name[n:n+1])
			if err != nil {
				println(err)
				os.Exit(1)
			}
			if name[n] == 0 {
				break
			}
			n++
		}
		fmt.Printf("%x%c\t%s\n", value, t)
		i += 5 + n + 1
	}
}
