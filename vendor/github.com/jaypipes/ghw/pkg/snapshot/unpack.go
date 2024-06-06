//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jaypipes/ghw/pkg/option"
)

const (
	TargetRoot = "ghw-snapshot-*"
)

const (
	// If set, `ghw` will not unpack the snapshot in the user-supplied directory
	// unless the aforementioned directory is empty.
	OwnTargetDirectory = 1 << iota
)

// Clanup removes the unpacket snapshot from the target root.
// Please not that the environs variable `GHW_SNAPSHOT_PRESERVE`, if set,
// will make this function silently skip.
func Cleanup(targetRoot string) error {
	if option.EnvOrDefaultSnapshotPreserve() {
		return nil
	}
	return os.RemoveAll(targetRoot)
}

// Unpack expands the given snapshot in a temporary directory managed by `ghw`. Returns the path of that directory.
func Unpack(snapshotName string) (string, error) {
	targetRoot, err := ioutil.TempDir("", TargetRoot)
	if err != nil {
		return "", err
	}
	_, err = UnpackInto(snapshotName, targetRoot, 0)
	return targetRoot, err
}

// UnpackInto expands the given snapshot in a client-supplied directory.
// Returns true if the snapshot was actually unpacked, false otherwise
func UnpackInto(snapshotName, targetRoot string, flags uint) (bool, error) {
	if (flags&OwnTargetDirectory) == OwnTargetDirectory && !isEmptyDir(targetRoot) {
		return false, nil
	}
	snap, err := os.Open(snapshotName)
	if err != nil {
		return false, err
	}
	defer snap.Close()
	return true, Untar(targetRoot, snap)
}

// Untar extracts data from the given reader (providing data in tar.gz format) and unpacks it in the given directory.
func Untar(root string, r io.Reader) error {
	var err error
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// we are done
			return nil
		}

		if err != nil {
			// bail out
			return err
		}

		if header == nil {
			// TODO: how come?
			continue
		}

		target := filepath.Join(root, header.Name)
		mode := os.FileMode(header.Mode)

		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, mode)
			if err != nil {
				return err
			}

		case tar.TypeReg:
			dst, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, mode)
			if err != nil {
				return err
			}

			_, err = io.Copy(dst, tr)
			if err != nil {
				return err
			}

			dst.Close()

		case tar.TypeSymlink:
			err = os.Symlink(header.Linkname, target)
			if err != nil {
				return err
			}
		}
	}
}

func isEmptyDir(name string) bool {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return false
	}
	return len(entries) == 0
}
