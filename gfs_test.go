package go_fuse_embed

import (
	"embed"
	"github.com/hanwen/go-fuse/v2/fuse"
	"os"
	"path/filepath"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
)

//go:embed testdata/*
var testFS embed.FS

func TestFuseEmbed(t *testing.T) {
	// Create a new FuseEmbed instance with the test filesystem
	fuseEmbed := New(&testFS, "testdata")

	// Create a temporary directory for mounting the filesystem
	mountDir, err := os.MkdirTemp("", "fusembed-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("Failed to remove temporary directory: %v", err)
		}
	}(mountDir)

	// Mount the filesystem
	server, err := fs.Mount(mountDir, fuseEmbed, &fs.Options{})
	if err != nil {
		t.Fatalf("Failed to mount filesystem: %v", err)
	}
	defer func(server *fuse.Server) {
		err := server.Unmount()
		if err != nil {
			t.Fatalf("Failed to unmount filesystem: %v", err)
		}
	}(server)

	// Wait for the filesystem to be ready
	err = server.WaitMount()
	if err != nil {
		t.Fatalf("Failed to wait for filesystem to be ready: %v", err)
	}

	// Verify the existence and content of files
	files := []string{
		"dir_a/file_a.txt",
		"dir_b/file_b.txt",
	}

	for _, file := range files {
		path := filepath.Join(mountDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read file '%s': %v", file, err)
			continue
		}
		expectedContent := "Content of " + filepath.Base(file)
		if string(content) != expectedContent {
			t.Errorf("Unexpected content for file '%s'. Expected: %s, Got: %s", file, expectedContent, string(content))
		}
	}
}
