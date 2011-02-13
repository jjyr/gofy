package main

import (
	"runtime"
	"unsafe"
)

const (
	BREAD = 1
	BDONE = 2
	BBUSY = 4
	BDELWRI = 8
	BASYNC = 16
)

type BlockMapper func(uint64) (uint64, Error)

type BlockDevice interface {
	DoEet(*Buf)
}

type Buf struct {
	Flags int
	Done, Want chan bool
	Count, Block uint64
	Data []byte
	Phys uint64
	Error
	*BIO
}

type BIO struct {
	all []Buf
	free chan *Buf
	BlockDevice
	BlockMapper
	BSize uint64
}

func (b *Buf) Release() {
	b.Flags &= ^(BBUSY | BASYNC)
	select {
	case b.BIO.free <- b:
	default:
		fuck("BIO free list full") // cleanup code would be more appropriate
	}
	for {
		select {
		case b.Want <- true:
		default:
			return
		}
	}
}

func (bio *BIO) Read(b uint64) (buf *Buf, err Error) {
	buf = bio.GetBuf(b)
	if buf.Flags & BDONE != 0 {
		return buf, nil
	}
	buf.Flags |= BREAD
	buf.Count = bio.BSize
	bio.BlockDevice.DoEet(buf)
	<- buf.Done
	buf.Flags |= BDONE
	return buf, buf.Error
}

func (b *Buf) WaitAndRelease() {
	<- b.Done
	b.Release()
}

func (b *Buf) Write() Error {
	b.Flags &= ^(BREAD | BDONE | BDELWRI)
	b.Error = nil
	b.Count = b.BSize
	b.BIO.BlockDevice.DoEet(b)
	if b.Flags & BASYNC == 0 {
		<- b.Done
		b.Release()
		return b.Error
	}
	go b.WaitAndRelease()
	return nil
}

func (b *Buf) DWrite() {
	b.Flags |= BDELWRI | BDONE
	b.Release()
}


func (bio *BIO) GetBuf(b uint64) (buf *Buf) {
again:
	for k := range bio.all {
		if bio.all[k].Block != b {
			continue
		}
		buf = &bio.all[k]
		if buf.Flags & BBUSY != 0 {
			<- buf.Want
			goto again
		}
		buf.Flags |= BBUSY
		return
	}
next:
	buf = <- bio.free 
	if buf.Flags & BBUSY != 0 {
		goto next
	}
	if buf.Flags & BDELWRI != 0 {
		buf.Flags |= BASYNC
		buf.Write()
		goto again
	}
	buf.Flags = BBUSY
	buf.Block = b
	buf.Error = nil
	return
}

func NewBIO(dev BlockDevice, N int, BSize uint64) (bio *BIO) {
	if (PAGESIZE % BSize) != 0 {
		fuck("BIO with non-divisors of the PAGESIZE not implemented")
	}
	bio = new(BIO)
	bio.BSize = BSize
	bio.BlockDevice = dev
	bio.free = make(chan *Buf, N)
	bio.all = make([]Buf, N)
	virt, phys := runtime.AlignAlloc(uint64(N) * BSize)
	for k := range bio.all {
		h := (*runtime.SliceHeader)(unsafe.Pointer(&bio.all[k].Data))
		h.Data, bio.all[k].Phys = virt, phys
		virt += uintptr(BSize)
		phys += BSize
		h.Len = int(BSize)
		h.Cap = int(BSize)

		bio.all[k].Done = make(chan bool)
		bio.all[k].Want = make(chan bool)
		bio.all[k].BIO = bio
		bio.all[k].Block = ^uint64(0)
		bio.free <- &bio.all[k]
	}
	return
}

// This produces (pseudo-) buffers for actually unbuffered I/O
// DON'T use those with the buffer functions, use DoEet
// DON'T use it in a large scale
// You have been warned
func MakeBuf(size uint64, dev BlockDevice) (b *Buf) {
	b = new(Buf)
	b.Done = make(chan bool)
	virt, phys := runtime.AlignAlloc(size)
	h := (*runtime.SliceHeader)(unsafe.Pointer(&b.Data))
	h.Data, b.Phys = virt, phys
	h.Len, h.Cap = int(size), int(size)
	b.BIO = new(BIO)
	b.BIO.BlockDevice = dev
	return
}

// This is for use with unbuffered buffers made by MakeBuf
// FIXME: this function does not free virtual memory
func (b *Buf) Free() {
	n := uint64((len(b.Data) + PAGESIZE - 1) / PAGESIZE)
	runtime.Ffree(b.Phys, n)
}
