package validators

import (
	"fmt"
	"strings"

	"github.com/plar/btrfs"
)

func ValidSubvolumeName(name string) error {
	if len(name) == 0 || strings.Index(name, "/") != -1 || name == "." || name == ".." {
		return fmt.Errorf("incorrect subvolume name '%s'", name)
	}

	if len(name) > btrfs.BtrfsVolNameMax {
		return fmt.Errorf("subvolume name too long '%s', max length is %d", name, btrfs.BtrfsVolNameMax)
	}

	return nil
}
