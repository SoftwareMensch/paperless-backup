package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"paperless-backup/internal/logger"
)

func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	// Create test directories with files
	dataDir := filepath.Join(tmpDir, "data")
	mediaDir := filepath.Join(tmpDir, "media")
	redisDir := filepath.Join(tmpDir, "redis")

	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(mediaDir, 0755)
	os.MkdirAll(redisDir, 0755)

	// Create some test files
	os.WriteFile(filepath.Join(dataDir, "file1.txt"), []byte("data content"), 0644)
	os.WriteFile(filepath.Join(mediaDir, "file2.txt"), []byte("media content"), 0644)
	os.WriteFile(filepath.Join(redisDir, "file3.txt"), []byte("redis content"), 0644)

	backupDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupDir, 0755)

	creator := New(log)
	backupFile := filepath.Join(backupDir, "test_backup.tar.gz")
	sourcePaths := []string{dataDir, mediaDir, redisDir}

	// Create backup
	err := creator.Create(backupFile, sourcePaths)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify backup file exists
	info, err := os.Stat(backupFile)
	if err != nil {
		t.Fatalf("Backup file should exist: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Backup file should not be empty")
	}

	// Check permissions
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestVerify(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	// Create a valid tar.gz file
	backupFile := filepath.Join(tmpDir, "test_backup.tar.gz")
	
	file, _ := os.Create(backupFile)
	gzWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzWriter)

	// Add a test file
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: 12,
	}
	tarWriter.WriteHeader(header)
	tarWriter.Write([]byte("test content"))

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	creator := New(log)

	// Verify should succeed
	err := creator.Verify(backupFile)
	if err != nil {
		t.Errorf("Verify should succeed for valid archive: %v", err)
	}
}

func TestVerifyInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	// Create an invalid tar.gz file
	backupFile := filepath.Join(tmpDir, "invalid.tar.gz")
	os.WriteFile(backupFile, []byte("not a valid tar.gz file"), 0644)

	creator := New(log)

	// Verify should fail
	err := creator.Verify(backupFile)
	if err == nil {
		t.Error("Verify should fail for invalid archive")
	}
}

func TestAddToTar(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	log, _ := logger.New(logPath)
	defer log.Close()

	// Create source directory with files
	sourceDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), []byte("content2"), 0644)

	// Create tar file
	tarFile := filepath.Join(tmpDir, "test.tar")
	file, _ := os.Create(tarFile)
	tarWriter := tar.NewWriter(file)

	creator := New(log)

	// Add directory to tar
	err := creator.addToTar(tarWriter, sourceDir)
	if err != nil {
		t.Fatalf("addToTar failed: %v", err)
	}

	tarWriter.Close()
	file.Close()

	// Verify tar contains files
	file, _ = os.Open(tarFile)
	defer file.Close()
	tarReader := tar.NewReader(file)

	fileCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error reading tar: %v", err)
		}
		fileCount++
		t.Logf("Found in tar: %s", header.Name)
	}

	if fileCount < 2 {
		t.Errorf("Expected at least 2 files in tar, got %d", fileCount)
	}
}

