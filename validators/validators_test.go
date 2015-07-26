package validators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidSubvolumeName(t *testing.T) {
	err := ValidSubvolumeName("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name ''")

	err = ValidSubvolumeName("/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name '/'")

	err = ValidSubvolumeName(".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name '.'")

	err = ValidSubvolumeName("..")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name '..'")

	err = ValidSubvolumeName(strings.Repeat("s", 512))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max length is 255")

	err = ValidSubvolumeName("subvol1")
	assert.NoError(t, err)
}
