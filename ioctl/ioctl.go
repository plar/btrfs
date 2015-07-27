package ioctl

/*
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ctree.h>
#include <btrfs/ioctl.h>
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"unsafe"
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

		var off C.size_t = 0

		for i := C.__u32(0); i < sk.nr_items; i++ {
			var item *C.struct_btrfs_root_item

			C.memcpy(unsafe.Pointer(&sh), unsafe.Pointer(&args.buf[off]), C.size_t(unsafe.Sizeof(sh)))
			off += C.size_t(unsafe.Sizeof(sh))
			item = (*C.struct_btrfs_root_item)(unsafe.Pointer(&args.buf[off]))
			off += C.size_t(sh.len)

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

func findUpdatedFiles(dir *C.DIR, rootId, oldestGen uint64) (uint64, error) {
	var maxFound uint64 = 0

	var args C.struct_btrfs_ioctl_search_args
	var sk *C.struct_btrfs_ioctl_search_key = &args.key
	var sh C.struct_btrfs_ioctl_search_header
	var item *C.struct_btrfs_file_extent_item
	var backup C.struct_btrfs_file_extent_item

	var foundGen = C.__u64(0)

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

		var off C.size_t = 0

		for i := C.__u32(0); i < sk.nr_items; i++ {
			C.memcpy(unsafe.Pointer(&sh), unsafe.Pointer(&args.buf[off]), C.size_t(unsafe.Sizeof(sh)))
			off += C.size_t(unsafe.Sizeof(sh))

			if sh.len == 0 {
				item = &backup
			} else {
				item = (*C.struct_btrfs_file_extent_item)(unsafe.Pointer(&args.buf[off]))
			}

			foundGen = C.__u64(item.generation)
			if sh._type == C.BTRFS_EXTENT_DATA_KEY && foundGen >= C.__u64(oldestGen) {
				// print...
			}

			off += C.size_t(sh.len)

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
