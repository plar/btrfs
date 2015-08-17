package ioctl

/*
#include <time.h>
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ctree.h>
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/pborman/uuid"
)

// ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
var overflow = errors.New("binary: varint overflows a X-bit integer")

func readU8(r io.ByteReader) (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return b, err
	}

	return uint8(b & 0xff), nil
}

func readU16(r io.ByteReader) (uint16, error) {
	var x uint16
	var s uint
	for i := 0; i < 2; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint16(b&0xff) << s
		s += 8
	}

	return x, nil
}

func readU32(r io.ByteReader) (uint32, error) {
	var x uint32
	var s uint
	for i := 0; i < 4; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint32(b&0xff) << s
		s += 8
	}

	return x, nil
}

func readU64(r io.ByteReader) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; i < 8; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		x |= uint64(b&0xff) << s
		s += 8
	}

	return x, nil
}

// func transid(data []byte, offs int64) (uint64, error) {
// 	r := bytes.NewReader(data)
// 	_, err := r.Seek(offs, 0)
// 	if err != nil {
// 		return 0, err
// 	}

// 	tid, err := binary.ReadUvarint(r)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return tid, nil
// }

func readUUID(r io.Reader) (uuid.UUID, error) {
	rawUUID := make([]byte, 16, 16)
	err := binary.Read(r, binary.LittleEndian, rawUUID)
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.UUID(rawUUID), nil
}

func addptr(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func hexdump(data []C.char, width int) {
	var hex []string
	var value []string
	for i := 0; i < len(data); i += width {
		h := fmt.Sprintf("%04X: ", i)
		var v string

		for j := 0; j < width; j++ {
			if i+j < len(data) {
				ch := (uint8)(data[i+j])
				h = h + fmt.Sprintf("%02X ", ch)
				if ch >= 32 && ch <= 127 {
					v = v + fmt.Sprintf("%s", string(data[i+j]))
				} else {
					v = v + fmt.Sprintf(".")
				}
			} else {
				h = h + ".. "
				v = v + " "
			}
		}
		if len(h) > 0 {
			hex = append(hex, h)
			value = append(value, v)
		}
	}

	for i := 0; i < len(hex); i++ {
		fmt.Printf("%s  %s\n", hex[i], value[i])
	}
}
