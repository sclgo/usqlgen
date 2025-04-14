package gen

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/ansel1/merry/v2"
	"github.com/murfffi/gorich/sclerr"
)

// This code is based on https://github.com/xyproto/unzip/blob/v1.0.0/unzip.go
// with multiple fixes on top. Original license is MIT.
// The code is copied instead of referenced because the original library pulled
// unnecessary dependencies for a CLI that we don't use.
// Note that original code didn't have unit tests so none we lost in the copy.

func ExtractZip(zipFilename, destPath string) error {

	// Open the source filename for reading
	zipReader, err := zip.OpenReader(zipFilename)
	if err != nil {
		return merry.Wrap(err)
	}
	defer sclerr.CloseQuietly(zipReader)

	// For each file in the archive
	for _, archiveReader := range zipReader.File {
		err = extractEntry(archiveReader, destPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractEntry(archiveReader *zip.File, destPath string) error {

	// Open the file in the archive
	archiveFile, err := archiveReader.Open()
	if err != nil {
		return merry.Wrap(err)
	}
	defer sclerr.CloseQuietly(archiveFile)

	// Prepare to write the file
	finalPath := filepath.Join(destPath, archiveReader.Name)

	// Check if the file to extract is just a directory
	if archiveReader.FileInfo().IsDir() {
		err = os.MkdirAll(finalPath, 0755)
		return err
	}

	// Create all needed directories
	if err = os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return err
	}

	// Prepare to write the destination file
	destinationFile, err := os.OpenFile(finalPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, archiveReader.Mode())
	if err != nil {
		return err
	}
	defer sclerr.CloseQuietly(destinationFile)

	// Write the destination file
	_, err = io.Copy(destinationFile, archiveFile)
	return err
}
