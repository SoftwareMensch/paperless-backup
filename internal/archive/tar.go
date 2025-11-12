package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"paperless-backup/internal/logger"
)

// Creator handles tar.gz archive creation and verification
type Creator struct {
	logger *logger.Logger
}

// New creates a new archive Creator
func New(logger *logger.Logger) *Creator {
	return &Creator{
		logger: logger,
	}
}

// Create creates a compressed tar.gz archive of specified paths
func (c *Creator) Create(outputPath string, sourcePaths []string) error {
	c.logger.Logf("INFO", "Creating compressed backup archive: %s", outputPath)

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer with no timestamp for deterministic output
	gzWriter := gzip.NewWriter(outFile)
	gzWriter.ModTime = time.Time{} // Zero time for reproducibility
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add each path to the tar
	for _, path := range sourcePaths {
		if err := c.addToTar(tarWriter, path); err != nil {
			return fmt.Errorf("failed to add %s to archive: %w", path, err)
		}
	}

	// Close writers to flush
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Set restrictive permissions
	if err := os.Chmod(outputPath, 0600); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Get backup size
	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	sizeMB := float64(info.Size()) / 1024 / 1024
	c.logger.Logf("INFO", "Backup created successfully: %s (%.2fMB)", outputPath, sizeMB)

	return nil
}

// addToTar recursively adds a directory and its contents to the tar archive
func (c *Creator) addToTar(tarWriter *tar.Writer, source string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update header name to use full path
		header.Name = path
		if strings.HasPrefix(header.Name, "/") {
			header.Name = strings.TrimPrefix(header.Name, "/")
		}

		// Normalize timestamps for deterministic archives
		// This ensures identical content produces identical checksums
		header.ModTime = time.Time{}
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If not a regular file (directory, symlink, etc.), skip content
		if !info.Mode().IsRegular() {
			return nil
		}

		// Write file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(tarWriter, file); err != nil {
			return err
		}

		return nil
	})
}

// Verify validates the integrity of a tar.gz archive
func (c *Creator) Verify(archivePath string) error {
	c.logger.Log("INFO", "Verifying backup integrity...")

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("backup integrity check failed (gzip): %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Read through all entries to verify integrity
	fileCount := 0
	for {
		_, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("backup integrity check failed (tar): %w", err)
		}
		fileCount++
	}

	c.logger.Logf("INFO", "Backup integrity check passed (%d files)", fileCount)
	return nil
}

