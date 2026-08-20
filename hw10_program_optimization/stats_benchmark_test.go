package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	reader, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer reader.Close()

	if len(reader.File) != 1 {
		b.Fatalf("expected one file in archive, got %d", len(reader.File))
	}

	file, err := reader.File[0].Open()
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	var data bytes.Buffer
	if _, err := data.ReadFrom(file); err != nil {
		b.Fatal(err)
	}

	input := data.Bytes()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := GetDomainStat(bytes.NewReader(input), "biz")
		if err != nil {
			b.Fatal(err)
		}
	}
}