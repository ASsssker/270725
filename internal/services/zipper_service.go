package services

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
)

type Zipper struct {
	archivePath string
}

func NewZipper(archivePath string) (*Zipper, error) {
	if err := os.MkdirAll(archivePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create zipper directory: %w", err)
	}

	return &Zipper{
		archivePath: archivePath,
	}, nil
}

func (z *Zipper) ToArchive(archiveName string, files map[string][]byte) error {
	archiveName = filepath.Join(filepath.Base(z.archivePath), archiveName)

	archive, err := os.Create(archiveName)
	if err != nil {
		return fmt.Errorf("failed to create zipper file: %w", err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	for filename, data := range files {
		w, err := zipWriter.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create zipWriter %w", err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("failed to write zipWriter %w", err)
		}
	}

	return nil
}
