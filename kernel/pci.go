package main

const (
	CONFIG_ADDRESS uint16 = 0xCF8
	CONFIG_DATA uint16 = 0xCFC
	NPCIDEVICES = 32
	NPCIBUS = 256

	PCICLASS_MASSSTORAGE uint8 = 1
	PCISUBCLASS_IDE uint8 = 1
)

type PCIDevice struct {
	Bus, Dev, Fun int
	VendID, DevID uint16
	Class, Subclass, ProgIF, RevID uint8
	Header uint8
	BAR [6]uint32
}

var pci []PCIDevice

func outl(uint16, uint32)
func inl(uint16) uint32

func ReadPCIConf(bus int, dev int, fun int, reg int) uint32 {
	outl(CONFIG_ADDRESS, (1<<31) | (uint32(bus) << 16) | (uint32(dev) << 11) | (uint32(fun) << 8) | (uint32(reg) << 2))
	return inl(CONFIG_DATA)
}

func scanbus(bus int) {
	for dev := 0; dev < NPCIDEVICES; dev++ {
		var p PCIDevice
		fun := 0
	nextfunc:
		x := ReadPCIConf(bus, dev, fun, 0)
		if (x & 0xFFFF) == 0xFFFF {
			continue
		}
		p.Bus = bus
		p.Dev = dev
		p.Fun = fun
		p.VendID = uint16(x)
		p.DevID = uint16(x >> 16)
		x = ReadPCIConf(bus, dev, fun, 2)
		p.RevID = uint8(x)
		p.ProgIF = uint8(x>>8)
		p.Subclass = uint8(x>>16)
		p.Class = uint8(x>>24)
		x = ReadPCIConf(bus, dev, fun, 3)
		p.Header = uint8(x >> 16)
		switch p.Header & 0x7f {
		case 0:
			for i := 0; i < 6; i++ {
				p.BAR[i] = ReadPCIConf(bus, dev, fun, 4+i)
			}
		}
		pci = append(pci, p)
		if p.Header & 0x80 != 0 {
			fun++
			if fun < 8 {
				goto nextfunc
			}
		}
	}
}

func initpci() {
	scanbus(0)
}
