package subvolume

import "github.com/plar/btrfs"

func init() {
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolCreate, ioctlCreate)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolSnapshot, ioctlSnapshot)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolFindNew, ioctlFindNew)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolDelete, ioctlDelete)

	btrfs.RegisterAPI(btrfs.CLI, btrfs.CmdSubvolCreate, cliCreate)
}
