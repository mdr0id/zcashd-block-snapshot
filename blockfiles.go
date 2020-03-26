package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func gzipDefaultBlocks(height int) error {
	file, err := os.Create("zcashd-" + strconv.Itoa(height) + ".tar.gz")
	if err != nil {
		return err
	}
	defer file.Close()
	// set up the gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// grab the paths that need to be added in
	paths := []string{
		"./blocks",
		"./chainstate",
	}

	// add each file as needed into the current tar archive
	for i := range paths {
		filepath.Walk(paths[i], func(file string, fileInfo os.FileInfo, err error) error {
			fmt.Printf("Adding file: %s\n", file)
			header, err := tar.FileInfoHeader(fileInfo, file)
			if err != nil {
				return err
			}
			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			header.Name = filepath.ToSlash(file)
			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fileInfo.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		},
		)
	}

	return nil
}
