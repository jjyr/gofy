package main

import "runtime"

type IDEDisk struct {
	*IDEController
	N int
	Avail bool
	Ident [512]byte
	Type, Capabilities uint16
	CommandSets, Cylinders, Heads, Sectors uint32
	Model string
	MaxLBA uint64
}

type IDEController struct{
	PCIDevice
	D [4]IDEDisk
	curdrive [2]int
	Handler chan *Buf
}

const (
	PDRTSIZE = 1
	
	IDEDATAPORT = 0
	IDEFEATURES = 1
	IDECOUNT = 2
	IDELBA0 = 3
	IDELBA1 = 4
	IDELBA2 = 5
	IDEDRIVE = 6
	IDECOMMAND = 7
	IDECONTROL = 12

	IDEREADSECTORS = 0x20
	IDEWRITESECTORS = 0x30
	IDECACHEFLUSH = 0xE7
	IDEIDENTIFY = 0xEC

	IDEBUSY = 0x80
	IDEDRQ = 8
	IDEDF = 1<<5
	IDEERR = 1

	IDELBA48 = 1<<26

	IDEBLOCKSIZE = 512

	NoSuchDriveError SimpleError = "No such drive"
	NoSuchBlockError SimpleError = "No such block"
	IOError SimpleError = "I/O error"
)

func outb(uint16, uint8)
func inb(uint16) uint8
func InPIO(uint16, []byte)
func OutPIO(uint16, []byte)

func (c *IDEDisk) getRegisterAddr(reg int) uint16 {
	if reg < 8 {
		r := uint16(c.IDEController.PCIDevice.BAR[c.N&2] & ^uint32(1)) + uint16(reg)
		return r
	}
	if reg == 12 {
		r := uint16(c.IDEController.PCIDevice.BAR[(c.N&2)+1] & ^uint32(1)) + 2
		return r
	}
	fuck("invalid register")
	return 0
}

func (c *IDEDisk) activate() {
	if c.IDEController.curdrive[c.N>>1] == c.N {
		return
	}
	c.writeRegister(IDEDRIVE, 0xA0 | 0x10 * (uint8(c.N) & 1))
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.IDEController.curdrive[c.N>>1] = c.N
}

func (c *IDEDisk) readRegister(reg int) uint8 {
	return inb(c.getRegisterAddr(reg))
}

func (c *IDEDisk) writeRegister(reg int, v uint8) {
	outb(c.getRegisterAddr(reg), v)
}

func (c *IDEDisk) identify(i int) bool {
	c.activate()
	c.writeRegister(IDECOUNT, 0)
	c.writeRegister(IDELBA0, 0)
	c.writeRegister(IDELBA1, 0)
	c.writeRegister(IDELBA2, 0)
	c.writeRegister(IDECOMMAND, IDEIDENTIFY)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	c.readRegister(IDECONTROL)
	if c.readRegister(IDECOMMAND) == 0 {
		return false
	}
/*	if c.readRegister(IDECOMMAND) == 0xFF {
		return false
	}*/
	for c.readRegister(IDECOMMAND) & IDEBUSY != 0 {
	}
	if c.readRegister(IDELBA0) != 0 && c.readRegister(IDELBA2) != 0 {
		return false
	}
	for c.readRegister(IDECOMMAND) & (IDEDRQ | IDEERR) == 0 {
	}
	if c.readRegister(IDECOMMAND) & IDEERR != 0 {
		return false
	}
	InPIO(c.getRegisterAddr(IDEDATAPORT), c.Ident[:])
	c.Avail = true
	c.Type = LE16(c.Ident[:])
	c.Cylinders = LE32(c.Ident[2:])
	c.Heads = LE32(c.Ident[6:])
	c.Sectors = LE32(c.Ident[10:])
	c.Capabilities = LE16(c.Ident[98:])
	c.CommandSets = LE32(c.Ident[164:])
	for i := 54; i < 94; i+=2 {
		c.Ident[i], c.Ident[i+1] = c.Ident[i+1], c.Ident[i]
	}
	last := 93
	for ; (c.Ident[last] == 0 || c.Ident[last] == ' ') && last >= 54; last-- {
	}
	c.Model = string(c.Ident[54:last+1])
	if c.CommandSets & IDELBA48 != 0 {
		c.MaxLBA = LE64(c.Ident[200:])
	} else {
		c.MaxLBA = uint64(LE32(c.Ident[120:]))
	}
	return true
}

func (c *IDEController) handler() {
	var block uint64

	for b := range c.Handler {
		disk, ok := b.BIO.BlockDevice.(*IDEDisk)
		b.Error = nil
		if !ok {
			b.Error = SimpleError("block with improper device passed to IDE driver")
			goto done
		}
		if !disk.Avail {
			b.Error = NoSuchDriveError
			goto done
		}
		if b.Block >= disk.MaxLBA || b.Block >= (1<<28) { // LBA48 not implemented
			b.Error = NoSuchBlockError
			goto done
		}
		if b.BlockMapper != nil {
			block, b.Error = b.BlockMapper(b.Block)
			if b.Error != nil {
				goto done
			}
		} else {
			block = b.Block
		}
		disk.activate()
		disk.writeRegister(IDEDRIVE, 0xE0 | uint8((disk.N & 1) << 4) | uint8((block >> 24) & 0xF))
		disk.writeRegister(IDECOUNT, uint8(b.BIO.BSize / IDEBLOCKSIZE))
		disk.writeRegister(IDELBA0, uint8(block))
		disk.writeRegister(IDELBA1, uint8(block >> 8))
		disk.writeRegister(IDELBA2, uint8(block >> 16))
		if b.Flags & BREAD != 0 {
			disk.writeRegister(IDECOMMAND, IDEREADSECTORS)
			for disk.readRegister(IDECOMMAND) & IDEBUSY != 0 {
				runtime.Gosched()
			}
			for disk.readRegister(IDECOMMAND) & (IDEDRQ | IDEERR | IDEDF) == 0 {
			}
			if disk.readRegister(IDECOMMAND) & (IDEDF | IDEERR) != 0 {
				b.Error = IOError
				goto done
			}
			InPIO(disk.getRegisterAddr(IDEDATAPORT), b.Data)
		} else {
			disk.writeRegister(IDECOMMAND, IDEWRITESECTORS)
			for disk.readRegister(IDECOMMAND) & IDEBUSY != 0 {
				runtime.Gosched()
			}
			for disk.readRegister(IDECOMMAND) & (IDEDRQ | IDEERR | IDEDF) == 0 {
			}
			if disk.readRegister(IDECOMMAND) & (IDEDF | IDEERR) != 0 {
				b.Error = IOError
				goto done
			}
			OutPIO(disk.getRegisterAddr(IDEDATAPORT), b.Data)
			disk.writeRegister(IDECOMMAND, IDECACHEFLUSH)
			for disk.readRegister(IDECOMMAND) & IDEBUSY != 0 {
			}
		}
	done:
		b.Done <- true
	}
}

func initide() (c *IDEController) {
	c = new(IDEController)
	for _, v := range pci {
		if v.Class == PCICLASS_MASSSTORAGE && v.Subclass == PCISUBCLASS_IDE && v.ProgIF & 0x80 != 0 {
			c.PCIDevice = v
			goto found
		}
	}
	fuck("No suitable IDE controller found!")
found:
	if c.PCIDevice.BAR[0] <= 1 {
		c.PCIDevice.BAR[0] = 0x1F0
	}
	if c.PCIDevice.BAR[1] <= 1 {
		c.PCIDevice.BAR[1] = 0x3F4
	}
	if c.PCIDevice.BAR[2] <= 1 {
		c.PCIDevice.BAR[2] = 0x170
	}
	if c.PCIDevice.BAR[3] <= 1 {
		c.PCIDevice.BAR[3] = 0x374
	}
	c.curdrive[0] = -1 // make sure the drive is really activated
	for i := 0; i < 2; i++ {
		c.D[i] = IDEDisk{IDEController: c, N: i}
		if c.D[i].identify(i) {
			print("ide", i, ": ", c.D[i].Model, ", ", c.D[i].MaxLBA >> 11, " MB \n")
		}
	}
	c.Handler = make(chan *Buf)
	go c.handler()
	return
}

func (c *IDEDisk) DoEet(b *Buf) {
	c.Handler <- b
}
