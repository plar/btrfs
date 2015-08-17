package ioctl

/*
#include <time.h>
#include <stdio.h>
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ctree.h>
#include <btrfs/ioctl.h>
#include <btrfs/btrfs-list.h>
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"unsafe"

	"github.com/pborman/uuid"
)

func free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

// TBD: Implement as OpenDirOrFile
func openDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer free(Cpath)

	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("Can't open dir '%s'", path)
	}
	return dir, nil
}

func closeDir(dir *C.DIR) {
	if dir != nil {
		C.closedir(dir)
	}
}

func getDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}

func SubvolCreate(path, name string) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_SUBVOL_CREATE,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to create btrfs subvolume: %v", errno.Error())
	}
	return nil
}

func SubvolSnapshot(readonly bool, src, dest, name string) error {
	srcDir, err := openDir(src)
	if err != nil {
		return err
	}
	defer closeDir(srcDir)

	destDir, err := openDir(dest)
	if err != nil {
		return err
	}
	defer closeDir(destDir)

	var args C.struct_btrfs_ioctl_vol_args_v2
	if readonly {
		args.flags |= C.BTRFS_SUBVOL_RDONLY
	}

	args.fd = C.__s64(getDirFd(srcDir))
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(destDir), C.BTRFS_IOC_SNAP_CREATE_V2,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to create btrfs snapshot: %v", errno.Error())
	}
	return nil
}

func SubvolDelete(path, name string) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_SNAP_DESTROY,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to destroy btrfs snapshot: %v", errno.Error())
	}
	return nil
}

func SubvolFindNew(name string, lastGen uint64) (uint64, error) {
	if ok, err := TestIsSubvolume(name); err != nil {
		return 0, err
	} else if !ok {
		return 0, fmt.Errorf("'%s' is not a subvolume", name)
	}

	subvolDir, err := openDir(name)
	if err != nil {
		return 0, err
	}
	defer closeDir(subvolDir)

	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(subvolDir), C.BTRFS_IOC_SYNC, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return 0, fmt.Errorf("Failed to fs-sync btrfs subvolume '%s': %v", name, errno.Error())
	}

	return findUpdatedFiles(subvolDir, 0, lastGen)
}

func TestIsSubvolume(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	stat := fi.Sys().(*syscall.Stat_t)

	// On btrfs subvolumes always have the inode 256
	return stat != nil && stat.Ino == 256 && fi.IsDir(), nil
}

func findRootGen(dir *C.DIR) (uint64, error) {
	fd := getDirFd(dir)

	var maxFound uint64 = 0
	var inoArgs C.struct_btrfs_ioctl_ino_lookup_args
	var args C.struct_btrfs_ioctl_search_args
	var sk *C.struct_btrfs_ioctl_search_key = &args.key
	var sh C.struct_btrfs_ioctl_search_header

	inoArgs.objectid = C.BTRFS_FIRST_FREE_OBJECTID

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BTRFS_IOC_INO_LOOKUP, uintptr(unsafe.Pointer(&inoArgs)))
	if errno != 0 {
		return 0, fmt.Errorf("Failed to perform the inode lookup %v", errno.Error())
	}

	sk.tree_id = 1
	sk.min_objectid = inoArgs.treeid
	sk.max_objectid = inoArgs.treeid
	sk.max_type = C.BTRFS_ROOT_ITEM_KEY
	sk.min_type = C.BTRFS_ROOT_ITEM_KEY
	sk.max_offset = math.MaxUint64
	sk.max_transid = math.MaxUint64
	sk.nr_items = 4096

	for {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BTRFS_IOC_TREE_SEARCH, uintptr(unsafe.Pointer(&args)))
		if errno != 0 {
			return 0, fmt.Errorf("Failed to perform the search %v", errno.Error())
		}

		if sk.nr_items == 0 {
			break
		}

		var off uintptr = 0

		for i := C.__u32(0); i < sk.nr_items; i++ {
			var item *C.struct_btrfs_root_item

			C.memcpy(unsafe.Pointer(&sh), addptr(unsafe.Pointer(&args.buf), off), C.sizeof_struct_btrfs_ioctl_search_header)
			off += C.sizeof_struct_btrfs_ioctl_search_header

			item = (*C.struct_btrfs_root_item)(unsafe.Pointer(&args.buf[off]))
			off += uintptr(sh.len)

			sk.min_objectid = sh.objectid
			sk.min_type = sh._type
			sk.min_offset = sh.offset

			if sh.objectid > inoArgs.treeid {
				break
			}

			if sh.objectid == inoArgs.treeid && sh._type == C.BTRFS_ROOT_ITEM_KEY {
				rootGeneration := item.generation
				if maxFound < uint64(rootGeneration) {
					maxFound = uint64(rootGeneration)
				}
			}
		}

		if sk.min_offset < math.MaxUint64 {
			sk.min_offset++
		} else {
			break
		}

		if sk.min_type != C.BTRFS_ROOT_ITEM_KEY {
			break
		}
		if sk.min_objectid != inoArgs.treeid {
			break
		}

	}

	return maxFound, nil
}

func findPathRootId(dir *C.DIR) (uint64, error) {
	fd := getDirFd(dir)

	var args C.struct_btrfs_ioctl_ino_lookup_args
	args.objectid = C.BTRFS_FIRST_FREE_OBJECTID

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BTRFS_IOC_INO_LOOKUP, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return 0, fmt.Errorf("Failed to perform the inode lookup %v", errno.Error())
	}

	return uint64(args.treeid), nil
}

func findUpdatedFiles(dir *C.DIR, rootId, oldestGen uint64) (uint64, error) {
	var maxFound uint64 = 0

	var args C.struct_btrfs_ioctl_search_args
	var sk *C.struct_btrfs_ioctl_search_key = &args.key
	var sh C.struct_btrfs_ioctl_search_header
	var item *BtrfsFileExtentItem
	var backup BtrfsFileExtentItem

	var foundGen uint64 = 0

	sk.tree_id = C.__u64(rootId)
	sk.max_objectid = math.MaxUint64
	sk.max_offset = math.MaxUint64
	sk.max_transid = math.MaxUint64
	sk.max_type = C.BTRFS_EXTENT_DATA_KEY
	sk.min_transid = C.__u64(oldestGen)
	sk.nr_items = 4096

	fd := getDirFd(dir)

	maxFound, err := findRootGen(dir)
	if err != nil {
		return 0, err
	}

	for {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BTRFS_IOC_TREE_SEARCH, uintptr(unsafe.Pointer(&args)))
		if errno != 0 {
			return 0, fmt.Errorf("Failed to perform the search %v", errno.Error())
		}

		if sk.nr_items == 0 {
			break
		}

		var off uintptr = 0

		for i := uint32(0); i < uint32(sk.nr_items); i++ {
			C.memcpy(unsafe.Pointer(&sh), addptr(unsafe.Pointer(&args.buf), off), C.sizeof_struct_btrfs_ioctl_search_header)
			off += C.sizeof_struct_btrfs_ioctl_search_header

			if sh.len == 0 {
				item = &backup
			} else {
				rawItem := (*C.struct_btrfs_file_extent_item)(addptr(unsafe.Pointer(&args.buf), off))
				item, err = NewBtrfsFileExtentItem(rawItem)
				if err != nil {
					return 0, err
				}
			}

			foundGen = item.Generation
			if sh._type == C.BTRFS_EXTENT_DATA_KEY && foundGen >= uint64(oldestGen) {
				// print...
			}

			off += uintptr(sh.len)

			sk.min_objectid = sh.objectid
			sk.min_offset = sh.offset
			sk.min_type = sh._type
		}

		sk.nr_items = 4096
		if sk.min_offset < math.MaxUint64 {
			sk.min_offset++
		} else if sk.min_objectid < math.MaxUint64 {
			sk.min_objectid++
			sk.min_offset = 0
			sk.min_type = 0

		} else {
			break
		}
	}

	return maxFound, nil
}

type SubvolSearchResult struct {
	Id           uint64
	Gen          uint64
	CGen         uint64
	Parent       uint64
	TopLevel     uint64
	OTime        BtrfsTimespec
	ParentUUID   uuid.UUID
	ReceivedUUID uuid.UUID
	UUID         uuid.UUID
	Path         string
}

func subvolSearch(dir *C.DIR) /**C.struct_root_lookup,*/ error {
	fd := getDirFd(dir)

	//fmt.Printf("Run list...: %v", fd)

	//var root_lookup *C.struct_root_lookup

	var args C.struct_btrfs_ioctl_search_args
	var sk *C.struct_btrfs_ioctl_search_key = &args.key
	var sh C.struct_btrfs_ioctl_search_header

	var ref *C.struct_btrfs_root_ref
	var ri *C.struct_btrfs_root_item

	var dirId uint64
	var ogen uint64
	var gen uint64
	var flags uint64
	var t uint64

	var ouuid, puuid, ruuid uuid.UUID
	emptyUUID := uuid.UUID{}

	/* search in the tree of tree roots */
	sk.tree_id = 1

	/*
	 * set the min and max to backref keys.  The search will
	 * only send back this type of key now.
	 */
	sk.max_type = C.BTRFS_ROOT_BACKREF_KEY
	sk.min_type = C.BTRFS_ROOT_ITEM_KEY

	sk.min_objectid = C.BTRFS_FIRST_FREE_OBJECTID

	/*
	 * set all the other params to the max, we'll take any objectid
	 * and any trans
	 */
	sk.max_objectid = C.BTRFS_LAST_FREE_OBJECTID & (1<<64 - 1)
	sk.max_offset = math.MaxUint64
	sk.max_transid = math.MaxUint64

	fmt.Printf("SK: %#v\n", sk)

	sk.nr_items = 4096

	for {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, C.BTRFS_IOC_TREE_SEARCH, uintptr(unsafe.Pointer(&args)))
		if errno != 0 {
			return fmt.Errorf("Failed to perform the search %v", errno.Error())
		}

		if sk.nr_items == 0 {
			break
		}

		var off uintptr = 0
		for i := uint32(0); i < uint32(sk.nr_items); i++ {

			C.memcpy(unsafe.Pointer(&sh), addptr(unsafe.Pointer(&args.buf), off), C.sizeof_struct_btrfs_ioctl_search_header)
			off += C.sizeof_struct_btrfs_ioctl_search_header

			if sh._type == C.BTRFS_ROOT_BACKREF_KEY {
				ref = (*C.struct_btrfs_root_ref)(addptr(unsafe.Pointer(&args.buf), off))
				goref, err := NewBtrfsRootRef(ref)
				if err != nil {
					return err
				}

				dirId = goref.DirId
				name := C.GoStringN((*C.char)(addptr(unsafe.Pointer(ref), C.sizeof_struct_btrfs_root_ref)), C.int(goref.NameLen))

				ssr := SubvolSearchResult{
					Id:       sh.objectid,
					Gen:      0,
					CGen:     0,
					Parent:   sh.offset,
					TopLevel: sh.offset, /* use ref_id? */

					/* empty */
					/*
						OTime
						UUID
						ParentUUID
						ReceivedUUID
					*/

					/*
						Path: TBD
					*/
				}

				fmt.Printf("objId: %v, offs: %v, dirId: %v, name: %v, nameLen: %v\n", sh.objectid, sh.offset, dirId, name, goref.NameLen)

			} else if sh._type == C.BTRFS_ROOT_ITEM_KEY {
				ri = (*C.struct_btrfs_root_item)(addptr(unsafe.Pointer(&args.buf), off))
				gori, err := NewBtrfsRootItem(ri)
				if err != nil {
					return err
				}

				//fmt.Printf("RI: %#v\n", gori)
				gen = gori.Generation
				flags = gori.Flags

				if sh.len > C.sizeof_struct_btrfs_root_item_v0 {

					t = gori.OTime.Sec
					ogen = gori.OTransId
					ouuid = gori.UUID
					puuid = gori.ParentUUID
					ruuid = gori.ReceivedUUID
				} else {
					ouuid = emptyUUID
					puuid = emptyUUID
					ruuid = emptyUUID
				}

				fmt.Printf("objId: %v, offs: %v, flags: %X, ogen: %v, gen: %v, time: %v, uuid: %s, puuid: %s, ruuid: %s\n",
					sh.objectid, sh.offset, flags, ogen, gen, t, ouuid, puuid, ruuid)
			}

			off += uintptr(sh.len)
			/*
			 * record the mins in sk so we can make sure the
			 * next search doesn't repeat this root
			 */
			sk.min_objectid = sh.objectid
			sk.min_type = sh._type
			sk.min_offset = sh.offset

		}

		sk.nr_items = 4096
		sk.min_offset++
		if sk.min_offset == 0 {
			sk.min_type++
		} else {
			continue
		}

		if sk.min_type > C.BTRFS_ROOT_BACKREF_KEY {
			sk.min_type = C.BTRFS_ROOT_ITEM_KEY
			sk.min_objectid++
		} else {
			continue
		}

		if sk.min_objectid > sk.max_objectid {
			break
		}
	}

	return nil
}

func SubvolList(name string, lastGen uint64) error {
	if ok, err := TestIsSubvolume(name); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("'%s' is not a subvolume", name)
	}

	subvolDir, err := openDir(name)
	if err != nil {
		return err
	}
	defer closeDir(subvolDir)

	// var args C.struct_btrfs_ioctl_vol_args
	// for i, c := range []byte(name) {
	// 	args.name[i] = C.char(c)
	// }
	// _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(subvolDir), C.BTRFS_IOC_SYNC, uintptr(unsafe.Pointer(&args)))
	// if errno != 0 {
	// 	return 0, fmt.Errorf("Failed to fs-sync btrfs subvolume '%s': %v", name, errno.Error())
	// }

	// return findUpdatedFiles(subvolDir, 0, lastGen)

	return subvolSearch(subvolDir)
}
