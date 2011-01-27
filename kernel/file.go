package main

const (
	ORD = 1
)

type Error interface {
	String() string
}

type SimpleError string
func (s SimpleError) String() string {
	return string(s)
}
const EOF SimpleError = "EOF"

type NotFoundError struct {}
func (NotFoundError) String() string {
	return "file not found"
}

type File interface {
	PRead([]byte, uint64) (uint64, Error)
	PWrite([]byte, uint64) (uint64, Error)
	Close()
}

type Filesystem interface {
	Open(name string, flags int, mode uint32) (File, Error)
}

type NamespaceEntry struct {
	string
	Filesystem
}
type Namespace []NamespaceEntry

func (ns Namespace) Open(name string, flags int, mode uint32) (File, Error) {
	for _, v := range ns {
		if v.string == name[:len(v.string)] {
			f, err := v.Filesystem.Open(name[len(v.string):], flags, mode)
			if _, ok := err.(NotFoundError); !ok {
				return f, nil
			}
		}
	}
	return nil, NotFoundError{}
}
