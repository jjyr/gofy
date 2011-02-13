package main

type Ext2SuperBlock struct {
	Inodes, Blocks, RBlocks, FBlocks, FInodes uint32
	Supernum, LogBlock, LogFrag, GroupBlocks, GroupFrags, GroupInodes uint32
	MountTime, WriteTime uint32
	MountCount, MaxMountCount, Magic, State, ErrorHandling, MinorVersion uint16
	LastCheck, Interval, OS, MajorVersion uint32
	RUID, RGID uint16
}

type Ext2FS struct {
	SuperBuf *Buf
	Super Ext2SuperBlock
}

func InitExt2(p Partition) (fs *Ext2FS, err Error) {
	fs = new(Ext2FS)
	fs.SuperBuf = MakeBuf(1024, p.BlockDevice)
	fs.SuperBuf.Flags |= BREAD
	fs.SuperBuf.Block = 0
//	fs.SuperBuf.BlockMapper = func(b uint64) (uint64, Error) {return p.Block(b)}
	fs.SuperBuf.Count = 1024
	p.DoEet(fs.SuperBuf)
	<- fs.SuperBuf.Done
	if fs.SuperBuf.Error != nil {
		return nil, fs.SuperBuf.Error
	}
	fs.Super.Inodes = LE32(fs.SuperBuf.Data[0:])
	fs.Super.Blocks = LE32(fs.SuperBuf.Data[4:])
	fs.Super.RBlocks = LE32(fs.SuperBuf.Data[8:])
	fs.Super.FBlocks = LE32(fs.SuperBuf.Data[12:])
	fs.Super.FInodes = LE32(fs.SuperBuf.Data[16:])
	fs.Super.Supernum = LE32(fs.SuperBuf.Data[20:])
	fs.Super.LogBlock = LE32(fs.SuperBuf.Data[24:])
	fs.Super.LogFrag = LE32(fs.SuperBuf.Data[28:])
	fs.Super.GroupBlocks = LE32(fs.SuperBuf.Data[32:])
	fs.Super.GroupFrags = LE32(fs.SuperBuf.Data[36:])
	fs.Super.GroupInodes = LE32(fs.SuperBuf.Data[40:])
	fs.Super.MountTime = LE32(fs.SuperBuf.Data[44:])
	fs.Super.WriteTime = LE32(fs.SuperBuf.Data[48:])
	fs.Super.MountCount = LE16(fs.SuperBuf.Data[52:])
	fs.Super.MaxMountCount = LE16(fs.SuperBuf.Data[54:])
	fs.Super.Magic = LE16(fs.SuperBuf.Data[56:])
	fs.Super.State = LE16(fs.SuperBuf.Data[58:])
	fs.Super.ErrorHandling = LE16(fs.SuperBuf.Data[60:])
	fs.Super.MinorVersion = LE16(fs.SuperBuf.Data[62:])
	fs.Super.LastCheck = LE32(fs.SuperBuf.Data[64:])
	fs.Super.Interval = LE32(fs.SuperBuf.Data[68:])
	fs.Super.OS = LE32(fs.SuperBuf.Data[72:])
	fs.Super.MajorVersion = LE32(fs.SuperBuf.Data[76:])
	fs.Super.RUID = LE16(fs.SuperBuf.Data[80:])
	fs.Super.RGID = LE16(fs.SuperBuf.Data[82:])
	println(hex(uint64(fs.SuperBuf.Data[0]), true))
	return
}
