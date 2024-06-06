//go:build !linux
// +build !linux

//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

func setupScratchDir(scratchDir string) error {
	return nil
}

func ExpectedCloneStaticContent() []string {
	return []string{}
}

func ExpectedCloneGPUContent() []string {
	return []string{}
}

func ExpectedCloneNetContent() []string {
	return []string{}
}

func ExpectedClonePCIContent() []string {
	return []string{}
}
