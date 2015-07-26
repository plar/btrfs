package validators

import "strings"

func IsSubvolumeName(name string) bool {
	return len(name) > 0 && strings.Index(name, "/") == -1 && name != "." && name != ".."
}
