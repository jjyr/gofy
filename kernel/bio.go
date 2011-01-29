package main

import (
	"runtime"
	"unsafe"
)

const (
	NBUF = 100
	BUFSIZE = 512

	BREAD = 1
	BDONE = 2
	BBUSY = 4
	BDELWRI = 8
	BASYNC = 16
)

type BlockDevice interface {
	DoEet(*Buf)
}

type Buf struct {
	Flags int
	Done, Want chan bool
	BlockDevice
	Count, Block uint64
	Data []byte
	Phys uint64
	Error
}

var (
	buffers [NBUF]Buf
	blocks [NBUF][BUFSIZE]byte // by placing this in the bss section, it will be identity mapped. magic.
	freebuf chan *Buf
)

func (b *Buf) Release() {
	for b.Want <- true {
	}
	if b.Error != nil {
		b.BlockDevice = nil
	}
	b.Flags &= ^(BBUSY | BASYNC)
	freebuf <- b
}

func BRead(d BlockDevice, b uint64) (buf *Buf, err Error) {
	buf = GetBuf(d, b)
	if buf.Flags & BDONE != 0 {
		return buf, nil
	}
	buf.Flags |= BREAD
	buf.Count = BUFSIZE
	d.DoEet(buf)
	<- buf.Done
	return buf, buf.Error
}

func (b *Buf) Write() (err Error) {
	b.Flags &= ^(BREAD | BDONE | BDELWRI)
	b.Error = nil
	b.Count = BUFSIZE
	b.BlockDevice.DoEet(b)
	if b.Flags & BASYNC == 0 {
		<- b.Done
		err = b.Error
		b.Release()
		return
	}
	return nil
}

func GetBuf(d BlockDevice, b uint64) (buf *Buf) {
again:
	if d != nil {
		for k := range buffers {
			if buffers[k].BlockDevice != d || buffers[k].Block != b {
				continue
			}
			buf = &buffers[k]
			if buf.Flags & BBUSY != 0 {
				<- buf.Want
				goto again
			}
			buf.Flags |= BBUSY
			return
		}
	}
next:
	buf = <- freebuf
	if buf.Flags & BBUSY != 0 {
		goto next
	}
	if buf.Flags & BDELWRI != 0 {
		buf.Flags |= BASYNC
		buf.Write()
		goto again
	}
	buf.Flags = BBUSY
	buf.BlockDevice = d
	buf.Block = b
	buf.Error = nil
	return
}

func initbio() {
	freebuf = make(chan *Buf, NBUF)
	for k := range buffers {
		buffers[k].Data = blocks[k][:]
		buffers[k].Done = make(chan bool)
		h := (*runtime.SliceHeader)(unsafe.Pointer(&buffers[k].Data))
		buffers[k].Phys = uint64(h.Data)
		freebuf <- &buffers[k]
	}
}
