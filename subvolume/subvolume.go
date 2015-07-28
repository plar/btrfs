package subvolume

import "github.com/plar/btrfs"

func init() {
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolumeCreate, ioctlCreate)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolumeSnapshot, ioctlSnapshot)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolumeFindNew, ioctlFindNew)
	btrfs.RegisterAPI(btrfs.IOCTL, btrfs.CmdSubvolumeDelete, ioctlDelete)

	btrfs.RegisterAPI(btrfs.CLI, btrfs.CmdSubvolumeCreate, cliCreate)
}
