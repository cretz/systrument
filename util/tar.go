package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// level is gzip constant
func CreateTarballFromDir(source string, out string, level int) error {
	file, err := os.Create(out)
	if err != nil {
		return err
	}
	gw, afterFileErr := gzip.NewWriterLevel(file, level)
	if afterFileErr == nil {
		tw := tar.NewWriter(gw)
		afterFileErr = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			unprefixed := strings.TrimPrefix(path, source)
			// Don't want the dir they sent us, just what's inside
			if unprefixed == "" {
				return nil
			}
			rel, err := filepath.Rel("/", strings.TrimPrefix(path, source))
			if err != nil {
				return fmt.Errorf("Unable to make relative path: %v", err)
			}
			// When building from Windows we need this to be nix-slashed
			header.Name = filepath.ToSlash(rel)
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		})
		tw.Close()
		gw.Close()
	}
	file.Close()
	return afterFileErr
}

func ExtractTarball(file *os.File, target string) error {
	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("Unable to read gzip: %v", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Error reading tar: %v", err)
		}
		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
		} else {
			childFile, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(childFile, tr)
			childFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
