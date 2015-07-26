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
	return fmt.Sprintf("qgroups=%v, ro='%s', src='%s', dest='%s', name='%s'", c.qgroups, c.readOnly, c.src, c.dest, c.name)
}

func (c *cmdSubvolSnapshot) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolumeSnapshot), Context: c.context(), Err: err}
}

func (c *cmdSubvolSnapshot) Execute() error {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlSnapshotExecute(c *cmdSubvolSnapshot) error {
	dest := filepath.Join(c.dest, c.name)
	fi, err := os.Stat(dest)
	if err == nil && !fi.IsDir() {
		return fmt.Errorf("'%s' exists and it is not a directory", dest)
	}

	var newname, dst string

	if fi.IsDir() {
		newname = filepath.Base(c.src)
		dst = c.dest
	} else {
		newname = filepath.Base(c.dest)
		dst = filepath.Dir(c.dest)
	}

	dest := filepath.Join(c.dest, c.name)
	if subvol, err := ioctl.TestIsSubvolume(dest); err != nil {
		return err
	} else if !subvol {
		return c.error(fmt.Errorf("'%s' is not a subvolume", dest))
	}

	err := validators.ValidSubvolumeName(newname)
	if err != nil {
		return err
	}

	err = ioctl.SubvolSnapshot(c.readOnly, c.src, c.dest, newname)
	if err != nil {
		return err
	}

	return nil
}

// btrfs cli executor
func cliSnapshotExecute(c *cmdSubvolSnapshot) error {
	err := c.validate()
	if err != nil {
		return err
	}

	return c.error(errors.New("Unimplemented"))
}

// commands
func ioctlSnapshot() interface{} {
	return &cmdSubvolSnapshot{executor: ioctlSnapshotExecute}
}

func cliSnapshot() interface{} {
	return &cmdSubvolSnapshot{executor: cliSnapshotExecute}
}
