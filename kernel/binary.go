package main

func LE16(b []byte) (r uint16) {
	r |= uint16(b[0])
	r |= uint16(b[1]) << 010
	return
}

func LE32(b []byte) (r uint32) {
	r |= uint32(b[0])
	r |= uint32(b[1]) << 010
	r |= uint32(b[2]) << 020
	r |= uint32(b[3]) << 030
	return
}

func LE64(b []byte) (r uint64) {
	r |= uint64(b[0])
	r |= uint64(b[1]) << 010
	r |= uint64(b[2]) << 020
	r |= uint64(b[3]) << 030
	r |= uint64(b[4]) << 040
	r |= uint64(b[5]) << 050
	r |= uint64(b[6]) << 060
	r |= uint64(b[7]) << 070
	return
}

func BE32(b []byte) (r uint32) {
	r |= uint32(b[0]) << 030
	r |= uint32(b[1]) << 020
	r |= uint32(b[2]) << 010
	r |= uint32(b[3])
	return
}
