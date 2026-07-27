package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	testCases := []struct {
		name    string
		offset  int64
		limit   int64
		wantErr error
	}{
		{
			name:   "copy whole file",
			offset: 0,
			limit:  0,
		},
		{
			name:   "copy first 10 bytes",
			offset: 0,
			limit:  10,
		},
		{
			name:   "copy first 1000 bytes",
			offset: 0,
			limit:  1000,
		},
		{
			name:   "limit exceeds file size",
			offset: 0,
			limit:  10000,
		},
		{
			name:   "offset 100 limit 1000",
			offset: 100,
			limit:  1000,
		},
		{
			name:   "offset 6000 limit 1000",
			offset: 6000,
			limit:  1000,
		},
		{
			name:    "offset exceeds file size",
			offset:  100000,
			limit:   0,
			wantErr: ErrOffsetExceedsFileSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			from := filepath.Join("testdata", "input.txt")
			to := filepath.Join(t.TempDir(), "output.txt")

			err := Copy(from, to, tc.offset, tc.limit)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := os.ReadFile(to)
			if err != nil {
				t.Fatal(err)
			}

			want, err := expectedContent(from, tc.offset, tc.limit)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Fatalf("copied file differs from expected")
			}
		})
	}
}

func expectedContent(path string, offset, limit int64) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if offset >= int64(len(data)) {
		return []byte{}, nil
	}

	data = data[offset:]

	if limit == 0 || limit > int64(len(data)) {
		return data, nil
	}

	return data[:limit], nil
}