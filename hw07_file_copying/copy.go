package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer src.Close()

	fileSize, err := validateFile(src, offset)
	if err != nil {
		return err
	}

	_, err = src.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	available := fileSize - offset
	bytesToCopy := getBytesToCopy(available, limit)

	if bytesToCopy == 0 {
		fmt.Println("100%")
		return nil
	}

	if err := copyData(bytesToCopy, src, dst); err != nil {
		return err
	}

	fmt.Println()

	return nil
}

func validateFile(src *os.File, offset int64) (int64, error) {
	info, err := src.Stat()
	if err != nil {
		return 0, err
	}

	if !info.Mode().IsRegular() {
		return 0, ErrUnsupportedFile
	}

	size := info.Size()

	if offset > size {
		return 0, ErrOffsetExceedsFileSize
	}

	return size, nil
}

func getBytesToCopy(available, limit int64) int64 {
	switch {
	case limit == 0:
		return available
	case limit > available:
		return available
	default:
		return limit
	}
}

func copyData(bytesToCopy int64, src *os.File, dst *os.File) error {
	buffer := make([]byte, 32*1024)

	var copied int64
	lastPercent := int64(-1)

	for copied < bytesToCopy {
		toRead := len(buffer)
		remaining := bytesToCopy - copied

		if remaining < int64(toRead) {
			toRead = int(remaining)
		}

		n, err := src.Read(buffer[:toRead])
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := dst.Write(buffer[:n]); err != nil {
			return err
		}

		copied += int64(n)

		percent := copied * 100 / bytesToCopy
		if percent != lastPercent {
			fmt.Printf("\r%d%%", percent)
			lastPercent = percent
		}
	}

	return nil
}
