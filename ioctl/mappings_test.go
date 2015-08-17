package ioctl

import (
	"bytes"
	"testing"

	"github.com/pborman/uuid"

	"github.com/stretchr/testify/assert"
)

type Dummy struct {
	D1 uint16
	D2 uint16
}

type Bar struct {
	D  Dummy
	B1 uint8
	B2 uint16
}

type Foo struct {
	F1   uint8
	F2   uint16
	F3   uint32
	F4   uint64
	BB   Bar
	F5   uint64 `seek:"0x100"`
	F6   uint64 `seek:"512"`
	F7   uint8
	UUID uuid.UUID
}

func TestNewStruct(t *testing.T) {
	data := make([]byte, 1024, 1024)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i + 1)
	}

	// test seek tag, hex value
	data[0x100] = 0x55
	data[0x101] = 0x66
	data[0x102] = 0x77
	data[0x103] = 0x88
	data[0x104] = 0x99
	data[0x105] = 0xAA
	data[0x106] = 0xBB
	data[0x107] = 0xCC

	// test seek tag, dec value
	data[512] = 0x11
	data[513] = 0x22
	data[514] = 0x33
	data[515] = 0x44
	data[516] = 0x55
	data[517] = 0x66
	data[518] = 0x77
	data[519] = 0x88

	f := &Foo{}
	err := NewStruct(f, bytes.NewReader(data[:]))
	assert.NoError(t, err)
	assert.Equal(t, uint8(0x01), f.F1)
	assert.Equal(t, uint16(0x0302), f.F2)
	assert.Equal(t, uint32(0x07060504), f.F3)
	assert.Equal(t, uint64(0x0f0e0d0c0b0a0908), f.F4)

	assert.Equal(t, uint16(0x1110), f.BB.D.D1)
	assert.Equal(t, uint16(0x1312), f.BB.D.D2)

	assert.Equal(t, uint8(0x14), f.BB.B1)
	assert.Equal(t, uint16(0x1615), f.BB.B2)

	assert.Equal(t, uint64(0xccbbaa9988776655), f.F5)
	assert.Equal(t, uint64(0x8877665544332211), f.F6)
	assert.Equal(t, uint8(0x09), f.F7)

	assert.Equal(t, uuid.UUID([]byte{0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19}), f.UUID)
}
