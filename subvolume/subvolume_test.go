package subvolume

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plar/btrfs"

	"github.com/stretchr/testify/assert"
)

var rootDir, mount string

func TestSubVolumeCreateValidation(t *testing.T) {
	//cmd := IoctlNewCreate()
	subvol := btrfs.NewIoctl().Subvolume()
	cmd := subvol.Create()
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name ''")

	cmd = subvol.Create()
	err = cmd.Name(strings.Repeat("s", 512)).Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subvolume name too long")
	assert.Contains(t, err.Error(), "max length is 255")

	cmd = subvol.Create()
	err = cmd.Name("/name").Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name '/name'")

	cmd = subvol.Create()
	err = cmd.Name(".").Execute()
	assert.Contains(t, err.Error(), "incorrect subvolume name '.'")

	cmd = subvol.Create()
	err = cmd.Name("..").Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect subvolume name '..'")
}

func TestSubVolumeCreate(t *testing.T) {
	subvol := btrfs.NewIoctl().Subvolume()
	cmd := subvol.Create()
	assert.NotNil(t, cmd)

	err := cmd.QuotaGroups("1", "2", "3").Destination(mount).Name("volume1").Execute()
	assert.NoError(t, err)

	ctx := cmd.(*cmdSubvolCreate)
	assert.Equal(t, ctx.qgroups, []string{"1", "2", "3"})
	assert.Equal(t, ctx.dest, mount)
	assert.Equal(t, ctx.name, "volume1")
}

func TestSubVolumeSnapshot(t *testing.T) {
	subvol := btrfs.NewIoctl().Subvolume()
	cmd := subvol.Snapshot()
	assert.NotNil(t, cmd)

	err := cmd.ReadOnly().Source("source").Destination(mount).Name("volume1").Execute()
	assert.NoError(t, err)

	ctx := cmd.(*cmdSubvolSnapshot)
	assert.Equal(t, ctx.readOnly, true)
	assert.Equal(t, ctx.src, "source")
	assert.Equal(t, ctx.dest, mount)
	assert.Equal(t, ctx.name, "volume1")
}

func run(cmd string, args ...string) error {
	log.Printf("Run %s %s", cmd, args)
	_, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func setup() {
	var err error

	rootDir, err = ioutil.TempDir("/var/tmp", "btrfs-test-")
	if err != nil {
		log.Fatalf("Cannot create tmp directory, err=%s", err)
	}

	mount = path.Join(rootDir, "btrfs")
	if err := os.MkdirAll(mount, 0700); err != nil {
		log.Fatalf("ERROR: MkdirAll %s, err=%s", mount, err)
	}

	imageFileName := filepath.Join(rootDir, "btrfs.img")
	ioutil.WriteFile(imageFileName, []byte("datadatadata"), 0700)
	os.Truncate(imageFileName, 1024*1024*1024) // 1GB

	if err := run("mkfs.btrfs", imageFileName); err != nil {
		log.Fatalf("ERROR: mkfs.btrfs %s, err=%s", imageFileName, err)
	}

	if err := run("mount", imageFileName, mount); err != nil {
		log.Fatalf("ERROR: mount %s %s, err=%s", imageFileName, mount, err)
	}
}

func teardown() {
	if err := run("umount", mount); err != nil {
		log.Fatal("ERROR: umount, err=%s", err)
	}

	// just to make sure that we're going to delete our temp directory
	if strings.HasPrefix(rootDir, "/var/tmp/btrfs-test-") {
		os.RemoveAll(rootDir)
	}

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
