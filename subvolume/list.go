package subvolume

import (
	"errors"
	"fmt"

	"github.com/plar/btrfs"
)

type subvolList struct {
	dest string

	executor func(c *subvolList) ([]btrfs.SubvolInfo, error)
}

func (c *subvolList) Path(dest string) btrfs.SubvolList {
	c.dest = dest
	return c
}

func (c *subvolList) FilterGeneration(filter string) btrfs.SubvolList {
	return c
}

func (c *subvolList) FilterOriginGeneration(filter string) btrfs.SubvolList {
	return c
}

func (c *subvolList) Sort(order string) btrfs.SubvolList {
	return c
}

func (c *subvolList) context() string {
	return fmt.Sprintf("dest='%s'", c.dest)
}

func (c *subvolList) error(err error) *btrfs.BtrfsError {
	return &btrfs.BtrfsError{Func: string(btrfs.CmdSubvolList), Context: c.context(), Err: err}
}

func (c *subvolList) Execute() ([]btrfs.SubvolInfo, error) {
	return c.executor(c)
}

// btrfs ioctl executor
func ioctlListExecute(c *subvolList) ([]btrfs.SubvolInfo, error) {
	if len(c.dest) == 0 {
		return nil, fmt.Errorf("Subvolume is required")
	}

	// path := filepath.Dir(c.dest)
	// name := filepath.Base(c.dest)

	// err := ioctl.SubvolList(path, name)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}

// btrfs cli executor
func cliListExecute(c *subvolList) ([]btrfs.SubvolInfo, error) {
	return nil, c.error(errors.New("Unimplemented"))
}

// commands
func ioctlList() interface{} {
	return &subvolList{executor: ioctlListExecute}
}

func cliList() interface{} {
	return &subvolList{executor: cliListExecute}
}
