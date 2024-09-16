//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package pcidb

import (
	"bufio"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	PCIIDS_URI = "https://pci-ids.ucw.cz/v2.2/pci.ids.gz"
	USER_AGENT = "golang-jaypipes-pcidb"
)

func (db *PCIDB) load(ctx *context) error {
	var foundPath string
	for _, fp := range ctx.searchPaths {
		if _, err := os.Stat(fp); err == nil {
			foundPath = fp
			break
		}
	}
	if foundPath == "" {
		if !ctx.enableNetworkFetch {
			return ERR_NO_DB
		}
		// OK, so we didn't find any host-local copy of the pci-ids DB file. Let's
		// try fetching it from the network and storing it
		if err := cacheDBFile(ctx.cachePath); err != nil {
			return err
		}
		foundPath = ctx.cachePath
	}
	f, err := os.Open(foundPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var scanner *bufio.Scanner
	if strings.HasSuffix(foundPath, ".gz") {
		var zipReader *gzip.Reader
		if zipReader, err = gzip.NewReader(f); err != nil {
			return err
		}
		defer zipReader.Close()
		scanner = bufio.NewScanner(zipReader)
	} else {
		scanner = bufio.NewScanner(f)
	}

	return parseDBFile(db, scanner)
}

func ensureDir(fp string) error {
	fpDir := filepath.Dir(fp)
	if _, err := os.Stat(fpDir); os.IsNotExist(err) {
		err = os.MkdirAll(fpDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// Pulls down the latest copy of the pci-ids file from the network and stores
// it in the local host filesystem
func cacheDBFile(cacheFilePath string) error {
	ensureDir(cacheFilePath)

	client := new(http.Client)
	request, err := http.NewRequest("GET", PCIIDS_URI, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", USER_AGENT)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	f, err := os.Create(cacheFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(cacheFilePath)
		}
	}()
	defer f.Close()
	// write the gunzipped contents to our local cache file
	zr, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}
	defer zr.Close()
	if _, err = io.Copy(f, zr); err != nil {
		return err
	}
	return err
}
