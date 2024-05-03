package go_fuse_embed

import (
	"context"
	"embed"
	gfs "github.com/hanwen/go-fuse/v2/fs"
	"io/fs"
	"path/filepath"
	"strings"
	"syscall"
)

var _ gfs.NodeOnAdder = (*FuseEmbed)(nil)

type FuseEmbed struct {
	gfs.Inode
	fs       *embed.FS
	prefix   string
	chmodMap map[string]uint32
}

func (f *FuseEmbed) OnAdd(ctx context.Context) {
	// Create a root inode
	root := &f.Inode

	// Variable to store the prefix to remove
	var prefix string

	// Iterate over the files in the embedded filesystem
	err := fs.WalkDir(f.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "." {
			return nil
		}

		if prefix == "" {
			prefix = f.prefix + "/"
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Remove the prefix from the path
		path = strings.TrimPrefix(path, prefix)

		// Split the path into directory and base components
		dir, base := filepath.Split(path)

		// Start from the root inode
		p := root

		// Add directories leading up to the file
		for _, component := range strings.Split(dir, "/") {
			if len(component) == 0 {
				continue
			}
			ch := p.GetChild(component)
			if ch == nil || ch == (*gfs.Inode)(nil) {
				// Create a directory
				ch = p.NewPersistentInode(ctx, &gfs.Inode{},
					gfs.StableAttr{Mode: syscall.S_IFDIR})
				// Add it
				p.AddChild(component, ch, true)
			}

			p = ch
		}

		// Read the file content
		content, err := fs.ReadFile(f.fs, prefix+path)
		if err != nil {
			return err
		}

		// Make a file out of the content bytes
		embedder := &gfs.MemRegularFile{
			Data: content,
		}

		// Set the file mode
		mode, ok := f.chmodMap[path]
		if ok {
			embedder.Attr.Mode = mode
		}

		// Create the file inode
		child := p.NewPersistentInode(ctx, embedder, gfs.StableAttr{})

		// Add the file to the parent directory
		p.AddChild(base, child, true)

		return nil
	})

	if err != nil {
		return
	}
}

func (f *FuseEmbed) ChmodFile(path string, mode uint32) {
	path = strings.TrimLeft(path, "/")
	f.chmodMap[path] = mode
}

func New(fs *embed.FS, prefix string) *FuseEmbed {
	return &FuseEmbed{
		fs:       fs,
		prefix:   prefix,
		chmodMap: make(map[string]uint32),
	}
}
