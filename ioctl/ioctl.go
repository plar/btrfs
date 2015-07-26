package ioctl

/*
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ioctl.h>
*/
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

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

func TestIsSubvolume(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	stat := fi.Sys().(*syscall.Stat_t)

	// On btrfs subvolumes always have the inode 256
	return stat != nil && stat.Ino == 256 && fi.IsDir(), nil
}
