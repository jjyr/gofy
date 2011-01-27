package main

type Initrd map[string] []byte

type InitrdFile []byte

func (f InitrdFile) PRead(targ []byte, off uint64) (n uint64, err Error) {
	n = uint64(len(targ))
	if uint64(len(f)) < n {
		n = uint64(len(f))
		err = EOF
	}
	copy(targ, f[off:])
	return
}

func (f InitrdFile) PWrite(targ []byte, off uint64) (n uint64, err Error) {
	return 0, SimpleError("invalid operation")
}

func (f InitrdFile) Close() {
}

func (fs Initrd) Open(name string, flags int, mode uint32) (File, Error) {
	if flags != ORD {
		return nil, SimpleError("invalid operation")
	}
	f, ok := fs[name]
	if !ok {
		return nil, NotFoundError{}
	}
	return InitrdFile(f), nil
}
