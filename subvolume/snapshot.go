package subvolume

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

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
	name     string

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

func (c *cmdSubvolSnapshot) Name(name string) btrfs.CmdSubvolSnapshot {
	c.name = name
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

func (c *cmdSubvolSnapshot) validate() error {
	c.name = strings.TrimSpace(c.name)
	if !validators.IsSubvolumeName(c.name) {
		return fmt.Errorf("incorrect subvolume name '%s'", c.name)
	}

	if len(c.name) > btrfs.BtrfsVolNameMax {
		return fmt.Errorf("subvolume name too long '%s', max length is %d", c.name, btrfs.BtrfsVolNameMax)
	}

	// dest := filepath.Join(c.dest, c.name)
	// fi, err := os.Stat(dest)
	// if err == nil && fi.IsDir() {
	// 	return fmt.Errorf("'%s' exists", dest)
	// }

	return nil
}

// btrfs ioctl executor
func ioctlSnapshotExecute(c *cmdSubvolSnapshot) error {
	err := c.validate()
	if err != nil {
		return err
	}

	dest := filepath.Join(c.dest, c.name)
	if subvol, err := ioctl.TestIsSubvolume(dest); err != nil {
		return err
	} else if !subvol {
		return c.error(fmt.Errorf("'%s' is not a subvolume", dest))
	}

	// err = ioctl.SubvolSnapshot(c.src, c.dest, c.name)
	// if err != nil {
	// 	return err
	// }

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
