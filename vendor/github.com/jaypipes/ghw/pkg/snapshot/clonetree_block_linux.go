//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func createBlockDevices(buildDir string) error {
	// Grab all the block device pseudo-directories from /sys/block symlinks
	// (excluding loopback devices) and inject them into our build filesystem
	// with all but the circular symlink'd subsystem directories
	devLinks, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return err
	}
	for _, devLink := range devLinks {
		dname := devLink.Name()
		if strings.HasPrefix(dname, "loop") {
			continue
		}
		devPath := filepath.Join("/sys/block", dname)
		trace("processing block device %q\n", devPath)

		// from the sysfs layout, we know this is always a symlink
		linkContentPath, err := os.Readlink(devPath)
		if err != nil {
			return err
		}
		trace("link target for block device %q is %q\n", devPath, linkContentPath)

		// Create a symlink in our build filesystem that is a directory
		// pointing to the actual device bus path where the block device's
		// information directory resides
		linkPath := filepath.Join(buildDir, "sys/block", dname)
		linkTargetPath := filepath.Join(
			buildDir,
			"sys/block",
			strings.TrimPrefix(linkContentPath, string(os.PathSeparator)),
		)
		trace("creating device directory %s\n", linkTargetPath)
		if err = os.MkdirAll(linkTargetPath, os.ModePerm); err != nil {
			return err
		}

		trace("linking device directory %s to %s\n", linkPath, linkContentPath)
		// Make sure the link target is a relative path!
		// if we use absolute path, the link target will be an absolute path starting
		// with buildDir, hence the snapshot will contain broken link.
		// Otherwise, the unpack directory will never have the same prefix of buildDir!
		if err = os.Symlink(linkContentPath, linkPath); err != nil {
			return err
		}
		// Now read the source block device directory and populate the
		// newly-created target link in the build directory with the
		// appropriate block device pseudofiles
		srcDeviceDir := filepath.Join(
			"/sys/block",
			strings.TrimPrefix(linkContentPath, string(os.PathSeparator)),
		)
		trace("creating device directory %q from %q\n", linkTargetPath, srcDeviceDir)
		if err = createBlockDeviceDir(linkTargetPath, srcDeviceDir); err != nil {
			return err
		}
	}
	return nil
}

func createBlockDeviceDir(buildDeviceDir string, srcDeviceDir string) error {
	// Populate the supplied directory (in our build filesystem) with all the
	// appropriate information pseudofile contents for the block device.
	devName := filepath.Base(srcDeviceDir)
	devFiles, err := ioutil.ReadDir(srcDeviceDir)
	if err != nil {
		return err
	}
	for _, f := range devFiles {
		fname := f.Name()
		fp := filepath.Join(srcDeviceDir, fname)
		fi, err := os.Lstat(fp)
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			// Ignore any symlinks in the deviceDir since they simply point to
			// either self-referential links or information we aren't
			// interested in like "subsystem"
			continue
		} else if fi.IsDir() {
			if strings.HasPrefix(fname, devName) {
				// We're interested in are the directories that begin with the
				// block device name. These are directories with information
				// about the partitions on the device
				buildPartitionDir := filepath.Join(
					buildDeviceDir, fname,
				)
				srcPartitionDir := filepath.Join(
					srcDeviceDir, fname,
				)
				trace("creating partition directory %s\n", buildPartitionDir)
				err = os.MkdirAll(buildPartitionDir, os.ModePerm)
				if err != nil {
					return err
				}
				err = createPartitionDir(buildPartitionDir, srcPartitionDir)
				if err != nil {
					return err
				}
			}
		} else if fi.Mode().IsRegular() {
			// Regular files in the block device directory are both regular and
			// pseudofiles containing information such as the size (in sectors)
			// and whether the device is read-only
			buf, err := ioutil.ReadFile(fp)
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					// example: /sys/devices/virtual/block/zram0/compact is 0400
					trace("permission denied reading %q - skipped\n", fp)
					continue
				}
				return err
			}
			targetPath := filepath.Join(buildDeviceDir, fname)
			trace("creating %s\n", targetPath)
			f, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err = f.Write(buf); err != nil {
				return err
			}
			f.Close()
		}
	}
	// There is a special file $DEVICE_DIR/queue/rotational that, for some hard
	// drives, contains a 1 or 0 indicating whether the device is a spinning
	// disk or not
	srcQueueDir := filepath.Join(
		srcDeviceDir,
		"queue",
	)
	buildQueueDir := filepath.Join(
		buildDeviceDir,
		"queue",
	)
	err = os.MkdirAll(buildQueueDir, os.ModePerm)
	if err != nil {
		return err
	}
	fp := filepath.Join(srcQueueDir, "rotational")
	buf, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	targetPath := filepath.Join(buildQueueDir, "rotational")
	trace("creating %s\n", targetPath)
	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	if _, err = f.Write(buf); err != nil {
		return err
	}
	f.Close()

	return nil
}

func createPartitionDir(buildPartitionDir string, srcPartitionDir string) error {
	// Populate the supplied directory (in our build filesystem) with all the
	// appropriate information pseudofile contents for the partition.
	partFiles, err := ioutil.ReadDir(srcPartitionDir)
	if err != nil {
		return err
	}
	for _, f := range partFiles {
		fname := f.Name()
		fp := filepath.Join(srcPartitionDir, fname)
		fi, err := os.Lstat(fp)
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			// Ignore any symlinks in the partition directory since they simply
			// point to information we aren't interested in like "subsystem"
			continue
		} else if fi.IsDir() {
			// The subdirectories in the partition directory are not
			// interesting for us. They have information about power events and
			// traces
			continue
		} else if fi.Mode().IsRegular() {
			// Regular files in the block device directory are both regular and
			// pseudofiles containing information such as the size (in sectors)
			// and whether the device is read-only
			buf, err := ioutil.ReadFile(fp)
			if err != nil {
				return err
			}
			targetPath := filepath.Join(buildPartitionDir, fname)
			trace("creating %s\n", targetPath)
			f, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err = f.Write(buf); err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}
