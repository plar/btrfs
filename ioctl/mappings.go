package ioctl

/**
 * Map BTRFS C structs to similar GoLang structs
 *
 * GoLang does not fully support aligned packed structures with "__attribute__ ((__packed__))" attribute
 * that is why using the native BTRFS structs via C.* interface leads to unpredictable results.
 *
 * To avoid this the reflect package has been used to write a mapping in general way from the raw data.
 * See NewStruct function for more details.
 */

/*
#include <time.h>
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ctree.h>
*/
import "C"

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/pborman/uuid"
)

type BtrfsTimespec struct {
	Sec  uint64
	NSec uint32
}

// struct btrfs_root_item {                                                     /*offs*/
//     struct btrfs_inode_item inode;                                           0x00
//     __le64 generation;                                                       0xa0
//     __le64 root_dirid;                                                       0xa8
//     __le64 bytenr;                                                           0xb0
//     __le64 byte_limit;                                                       0xb8
//     __le64 bytes_used;                                                       0xc0
//     __le64 last_snapshot;                                                    0xc8
//     __le64 flags;                                                            0xd0
//     __le32 refs;                                                             0xd8
//     struct btrfs_disk_key drop_progress;                                     0xdc
//     u8 drop_level;                                                           0xed
//     u8 level;                                                                0xee
//     __le64 generation_v2;                                                    0xef
//     u8 uuid[BTRFS_UUID_SIZE];                                                0xf7
//     u8 parent_uuid[BTRFS_UUID_SIZE];                                         0x107
//     u8 received_uuid[BTRFS_UUID_SIZE];                                       0x117
//     __le64 ctransid; /* updated when an inode changes */                     0x127
//     __le64 otransid; /* trans when created */                                0x12f
//     __le64 stransid; /* trans when sent. non-zero for received subvol */     0x137
//     __le64 rtransid; /* trans when received. non-zero for received subvol */ 0x13f
//     struct btrfs_timespec ctime;                                             0x147
//     struct btrfs_timespec otime;                                             0x153
//     struct btrfs_timespec stime;                                             0x15f
//     struct btrfs_timespec rtime;                                             0x16b
//     __le64 reserved[8]; /* for future */                                     0x177
// } __attribute__ ((__packed__));

type BtrfsRootItem struct {
	Generation   uint64 `seek:"0xA0"`
	RootDirId    uint64
	ByteNr       uint64
	ByteLimit    uint64
	BytesUsed    uint64
	LastSnapshot uint64
	Flags        uint64
	Refs         uint32

	DropLevel    uint8 `seek:"0xED"`
	Level        uint8
	GenerationV2 uint64
	UUID         uuid.UUID
	ParentUUID   uuid.UUID
	ReceivedUUID uuid.UUID
	CTransId     uint64
	OTransId     uint64
	STransId     uint64
	RTransId     uint64
	CTime        BtrfsTimespec
	OTime        BtrfsTimespec
	STime        BtrfsTimespec
	RTime        BtrfsTimespec
}

type BtrfsRootRef struct {
	DirId    uint64
	Sequence uint64
	NameLen  uint16
}

// 874 struct btrfs_file_extent_item {
// 875         /*
// 876          * transaction id that created this extent
// 877          */
// 878         __le64 generation;
// 879         /*
// 880          * max number of bytes to hold this extent in ram
// 881          * when we split a compressed extent we can't know how big
// 882          * each of the resulting pieces will be.  So, this is
// 883          * an upper limit on the size of the extent in ram instead of
// 884          * an exact limit.
// 885          */
// 886         __le64 ram_bytes;
// 887
// 888         /*
// 889          * 32 bits for the various ways we might encode the data,
// 890          * including compression and encryption.  If any of these
// 891          * are set to something a given disk format doesn't understand
// 892          * it is treated like an incompat flag for reading and writing,
// 893          * but not for stat.
// 894          */
// 895         u8 compression;
// 896         u8 encryption;
// 897         __le16 other_encoding; /* spare for later use */
// 898
// 899         /* are we inline data or a real extent? */
// 900         u8 type;
// 901
// 902         /*
// 903          * disk space consumed by the extent, checksum blocks are included
// 904          * in these numbers
// 905          *
// 906          * At this offset in the structure, the inline extent data start.
// 907          */
// 908         __le64 disk_bytenr;
// 909         __le64 disk_num_bytes;
// 910         /*
// 911          * the logical offset in file blocks (no csums)
// 912          * this extent record is for.  This allows a file extent to point
// 913          * into the middle of an existing extent on disk, sharing it
// 914          * between two snapshots (useful if some bytes in the middle of the
// 915          * extent have changed
// 916          */
// 917         __le64 offset;
// 918         /*
// 919          * the logical number of file blocks (no csums included).  This
// 920          * always reflects the size uncompressed and without encoding.
// 921          */
// 922         __le64 num_bytes;
// 923
// 924 } __attribute__ ((__packed__));

type BtrfsFileExtentItem struct {
	Generation    uint64
	RamBytes      uint64
	Compression   uint8
	Encryption    uint8
	OtherEncoding uint16
	Type          uint8
	DiskByteNr    uint64
	DiskNumBytes  uint64
	Offset        uint64
	NumBytes      uint64
}

func NewBtrfsRootItem(s *C.struct_btrfs_root_item) (*BtrfsRootItem, error) {
	raw := unsafe.Pointer(s)
	data := *(*[C.sizeof_struct_btrfs_root_item]byte)(raw)
	r := bytes.NewReader(data[:])

	var ri *BtrfsRootItem = &BtrfsRootItem{}
	err := NewStruct(ri, r)
	return ri, err
}

func NewBtrfsRootRef(s *C.struct_btrfs_root_ref) (*BtrfsRootRef, error) {
	raw := unsafe.Pointer(s)
	data := *(*[C.sizeof_struct_btrfs_root_ref]byte)(raw)
	r := bytes.NewReader(data[:])

	var rr *BtrfsRootRef = &BtrfsRootRef{}
	err := NewStruct(rr, r)
	return rr, err

}

func NewBtrfsFileExtentItem(s *C.struct_btrfs_file_extent_item) (*BtrfsFileExtentItem, error) {
	raw := unsafe.Pointer(s)
	data := *(*[C.sizeof_struct_btrfs_file_extent_item]byte)(raw)
	r := bytes.NewReader(data[:])

	var fei *BtrfsFileExtentItem = &BtrfsFileExtentItem{}
	err := NewStruct(fei, r)
	return fei, err
}

func NewStruct(dest interface{}, r io.ByteReader) error {

	value := reflect.ValueOf(dest).Elem()
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		sfl := value.Field(i)
		fl := typ.Field(i)
		rawSeek := fl.Tag.Get("seek")
		if len(rawSeek) > 0 {
			base := 10
			if strings.HasPrefix(rawSeek, "0x") {
				base = 16
				rawSeek = rawSeek[2:]
			}
			seek, err := strconv.ParseUint(rawSeek, base, 32)
			if err != nil {
				return err
			}

			seeker, ok := r.(io.Seeker)
			if !ok {
				return errors.New("io.Seeker interface is required")
			}
			seeker.Seek(int64(seek), 0)
		}

		if sfl.IsValid() && sfl.CanSet() {
			var err error
			switch fl.Type.Kind() {
			case reflect.Uint8:
				u8, err := readU8(r)
				if err == nil {
					sfl.SetUint(uint64(u8))
				}
			case reflect.Uint16:
				u16, err := readU16(r)
				if err == nil {
					sfl.SetUint(uint64(u16))
				}
			case reflect.Uint32:
				u32, err := readU32(r)
				if err == nil {
					sfl.SetUint(uint64(u32))
				}
			case reflect.Uint64:
				u64, err := readU64(r)
				if err == nil {
					sfl.SetUint(u64)
				}

			case reflect.Slice:
				if sfl.Type() == reflect.TypeOf(uuid.UUID{}) {
					uuidValue, err := readUUID(r.(io.Reader))
					if err == nil {
						sfl.SetBytes([]byte(uuidValue))
					}
				} else {

				}

			case reflect.Struct:
				err = NewStruct(sfl.Addr().Interface(), r)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}
