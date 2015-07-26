package subvolume

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plar/btrfs"
	"github.com/plar/btrfs/ioctl"
	"github.com/plar/btrfs/validators"
)

// snapshot command
type cmdSubvolSnapshot struct {
	qgroups  []string
	readOnly bool
	src      string
	dest     string

	executor func(c *cmdSubvolSnapshot) error
}

func (c *cmdSubvolSnapshot) QuotaGroups(qgroups ...string) btrfs.CmdSubvolSnapshot {
	for _, qgroup := range qgroups {
		c.qgroups = append(c.qgroups, qgroup)
	}
	return c
}

func (c *cmdSubvolSnapshot) ReadOnly() btrfs.CmdSubvolSnapshot {
	c.readOnly = true
	return c
}

func (c *cmdSubvolSnapshot) Source(src string) btrfs.CmdSubvolSnapshot {
	c.src = src
	return c
}

func (c *cmdSubvolSnapshot) Destination(dest string) btrfs.CmdSubvolSnapshot {
	c.dest = dest
	return c
}

func (c *cmdSubvolSnapshot) context() string {
	return fmt.Sprintf("qgroups=%v, ro='%s', src='%s', dest='%s'", c.qgroups, c.readOnly, c.src, c.dest)
}

func (c *cmdSubvolSnapshot) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeSnapshot), Context: c.context(), Err: err}
}

func (c *cmdSubvolSnapshot) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlSnapshotExecute(c *cmdSubvolSnapshot) error {
	fi, err := os.Stat(c.dest)
	if err == nil && !fi.IsDir() {
		return fmt.Errorf("'%s' exists and it is not a directory", c.dest)
	}

	var newname, dest string
	if err == nil && fi.IsDir() {
		dest = c.dest
		newname = filepath.Base(c.src)
	} else {
		dest = filepath.Dir(c.dest)
		newname = filepath.Base(c.dest)
	}

	if subvol, err := ioctl.TestIsSubvolume(c.src); err != nil {
		return err
	} else if !subvol {
		return c.error(fmt.Errorf("'%s' is not a subvolume", newname))
	}

	err = validators.ValidSubvolumeName(newname)
	if err != nil {
		return err
	}

	err = ioctl.SubvolSnapshot(c.readOnly, c.src, dest, newname)
	if err != nil {
		return err
	}

	return nil
}

// btrfs cli executor
func cliSnapshotExecute(c *cmdSubvolSnapshot) error {
	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlSnapshot() interface{} {
	return &cmdSubvolSnapshot{executor: ioctlSnapshotExecute}
}

func cliSnapshot() interface{} {
	return &cmdSubvolSnapshot{executor: cliSnapshotExecute}
}
