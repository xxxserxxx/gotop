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

// Attempting to tar up pseudofiles like /proc/cpuinfo is an exercise in
// futility. Notably, the pseudofiles, when read by syscalls, do not return the
// number of bytes read. This causes the tar writer to write zero-length files.
//
// Instead, it is necessary to build a directory structure in a tmpdir and
// create actual files with copies of the pseudofile contents

// CloneTreeInto copies all the pseudofiles that ghw will consume into the root
// `scratchDir`, preserving the hieratchy.
func CloneTreeInto(scratchDir string) error {
	err := setupScratchDir(scratchDir)
	if err != nil {
		return err
	}
	fileSpecs := ExpectedCloneContent()
	return CopyFilesInto(fileSpecs, scratchDir, nil)
}

// ExpectedCloneContent return a slice of glob patterns which represent the pseudofiles
// ghw cares about.
// The intended usage of this function is to validate a clone tree, checking that the
// content matches the expectations.
// Beware: the content is host-specific, because the content pertaining some subsystems,
// most notably PCI, is host-specific and unpredictable.
func ExpectedCloneContent() []string {
	fileSpecs := ExpectedCloneStaticContent()
	fileSpecs = append(fileSpecs, ExpectedCloneNetContent()...)
	fileSpecs = append(fileSpecs, ExpectedClonePCIContent()...)
	fileSpecs = append(fileSpecs, ExpectedCloneGPUContent()...)
	return fileSpecs
}

// ValidateClonedTree checks the content of a cloned tree, whose root is `clonedDir`,
// against a slice of glob specs which must be included in the cloned tree.
// Is not wrong, and this functions doesn't enforce this, that the cloned tree includes
// more files than the necessary; ghw will just ignore the files it doesn't care about.
// Returns a slice of glob patters expected (given) but not found in the cloned tree,
// and the error during the validation (if any).
func ValidateClonedTree(fileSpecs []string, clonedDir string) ([]string, error) {
	missing := []string{}
	for _, fileSpec := range fileSpecs {
		matches, err := filepath.Glob(filepath.Join(clonedDir, fileSpec))
		if err != nil {
			return missing, err
		}
		if len(matches) == 0 {
			missing = append(missing, fileSpec)
		}
	}
	return missing, nil
}

// CopyFileOptions allows to finetune the behaviour of the CopyFilesInto function
type CopyFileOptions struct {
	// IsSymlinkFn allows to control the behaviour when handling a symlink.
	// If this hook returns true, the source file is treated as symlink: the cloned
	// tree will thus contain a symlink, with its path adjusted to match the relative
	// path inside the cloned tree. If return false, the symlink will be deferred.
	// The easiest use case of this hook is if you want to avoid symlinks in your cloned
	// tree (having duplicated content). In this case you can just add a function
	// which always return false.
	IsSymlinkFn func(path string, info os.FileInfo) bool
	// ShouldCreateDirFn allows to control if empty directories listed as clone
	// content should be created or not. When creating snapshots, empty directories
	// are most often useless (but also harmless). Because of this, directories are only
	// created as side effect of copying the files which are inside, and thus directories
	// are never empty. The only notable exception are device driver on linux: in this
	// case, for a number of technical/historical reasons, we care about the directory
	// name, but not about the files which are inside.
	// Hence, this is the only case on which ghw clones empty directories.
	ShouldCreateDirFn func(path string, info os.FileInfo) bool
}

// CopyFilesInto copies all the given glob files specs in the given `destDir` directory,
// preserving the directory structure. This means you can provide a deeply nested filespec
// like
// - /some/deeply/nested/file*
// and you DO NOT need to build the tree incrementally like
// - /some/
// - /some/deeply/
// ...
// all glob patterns supported in `filepath.Glob` are supported.
func CopyFilesInto(fileSpecs []string, destDir string, opts *CopyFileOptions) error {
	if opts == nil {
		opts = &CopyFileOptions{
			IsSymlinkFn:       isSymlink,
			ShouldCreateDirFn: isDriversDir,
		}
	}
	for _, fileSpec := range fileSpecs {
		trace("copying spec: %q\n", fileSpec)
		matches, err := filepath.Glob(fileSpec)
		if err != nil {
			return err
		}
		if err := copyFileTreeInto(matches, destDir, opts); err != nil {
			return err
		}
	}
	return nil
}

func copyFileTreeInto(paths []string, destDir string, opts *CopyFileOptions) error {
	for _, path := range paths {
		trace("  copying path: %q\n", path)
		baseDir := filepath.Dir(path)
		if err := os.MkdirAll(filepath.Join(destDir, baseDir), os.ModePerm); err != nil {
			return err
		}

		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}
		// directories must be listed explicitly and created separately.
		// In the future we may want to expose this decision as hook point in
		// CopyFileOptions, when clear use cases emerge.
		destPath := filepath.Join(destDir, path)
		if fi.IsDir() {
			if opts.ShouldCreateDirFn(path, fi) {
				if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
					return err
				}
			} else {
				trace("expanded glob path %q is a directory - skipped\n", path)
			}
			continue
		}
		if opts.IsSymlinkFn(path, fi) {
			trace("    copying link: %q -> %q\n", path, destPath)
			if err := copyLink(path, destPath); err != nil {
				return err
			}
		} else {
			trace("    copying file: %q -> %q\n", path, destPath)
			if err := copyPseudoFile(path, destPath); err != nil && !errors.Is(err, os.ErrPermission) {
				return err
			}
		}
	}
	return nil
}

func isSymlink(path string, fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

func isDriversDir(path string, fi os.FileInfo) bool {
	return strings.Contains(path, "drivers")
}

func copyLink(path, targetPath string) error {
	target, err := os.Readlink(path)
	if err != nil {
		return err
	}
	trace("      symlink %q -> %q\n", target, targetPath)
	if err := os.Symlink(target, targetPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}

	return nil
}

func copyPseudoFile(path, targetPath string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
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
