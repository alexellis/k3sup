// Copyright (c) arkade author(s) 2022. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.
package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Unzip reads the compressed zip file from reader and writes it into dir.
// Unzip works similar to Untar where support for nested folders is removed
// so that all files are placed in the same target directory
func Unzip(reader io.ReaderAt, size int64, dir string, quiet bool) error {
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return fmt.Errorf("error creating zip reader: %s", err)
	}

	return unzip(*zipReader, dir, quiet)
}

func unzip(r zip.Reader, dir string, quiet bool) (err error) {
	if err != nil {
		return err
	}

	t0 := time.Now()
	nFiles := 0
	madeDir := map[string]bool{}
	defer func() {
		td := time.Since(t0)
		if err == nil {
			if !quiet {
				log.Printf("extracted zip into %s: %d files, %d dirs (%v)", dir, nFiles, len(madeDir), td)
			}
		} else {
			log.Printf("error extracting zip into %s after %d files, %d dirs, %v: %v", dir, nFiles, len(madeDir), td, err)
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		baseFile := filepath.Base(f.Name)
		abs := path.Join(dir, baseFile)

		if !quiet {
			fmt.Printf("Extracting: %s\n", abs)
		}

		fi := f.FileInfo()
		mode := fi.Mode()

		switch {
		case mode.IsDir():
			break
		case mode.IsRegular():
			f, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		default:
		}
		nFiles++
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
