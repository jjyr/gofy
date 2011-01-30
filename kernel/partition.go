package main

type Partition struct {
	Start, End uint64
}

func (p *Partition) Block(b uint64) (uint64, Error) {
	if b > p.End {
		return 0, SimpleError("no such block")
	}
	return b + p.Start, nil
}

func ReadMBR(b BlockDevice) (ps []Partition, err Error) {
	buf := MakeBuf(512, b)
	defer buf.Free()
	
	buf.Count = 512
	buf.Flags |= BREAD
	b.DoEet(buf)
	<- buf.Done
	if buf.Error != nil {
		return nil, buf.Error
	}
	for i := 0; i < 4; i++ {
		part := buf.Data[446 + 16 * i : 446 + 16 * (i + 1)]
		if part[0] & 0x7F != 0 {
			continue
		}
		if part[4] == 0 {
			continue
		}
		start := uint64(LE32(part[0x08:]))
		len := uint64(LE32(part[0x0C:]))
		if len == 0 {
			continue
		}
		ps = append(ps, Partition{Start: start, End: start + len})
	}
	return ps, nil
}
