// Package fsentry
//
// Allows storing hierarchical data in files and folders on the file system
// with json descriptions and creation/update timestamps.
package fsentry

import (
	binaryService "github.com/HardDie/fsentry/internal/binary/service"
	entryService "github.com/HardDie/fsentry/internal/entry/service"
	folderService "github.com/HardDie/fsentry/internal/folder/service"
	fsStorage "github.com/HardDie/fsentry/internal/fs/storage"
	"github.com/HardDie/fsentry/internal/service"
	"github.com/HardDie/fsentry/pkg/fsentry"
)

type Config struct {
	log      fsentry.Logger
	root     string
	isPretty bool
}

func WithLogger(log fsentry.Logger) func(cfg *Config) {
	return func(cfg *Config) {
		if log == nil {
			return
		}
		cfg.log = log
	}
}
func WithPretty() func(cfg *Config) {
	return func(cfg *Config) {
		cfg.isPretty = true
	}
}

func NewFSEntry(root string, ops ...func(fs *Config)) fsentry.IFSEntry {
	cfg := &Config{
		root: root,
	}
	for _, op := range ops {
		op(cfg)
	}

	fileStorage := fsStorage.New()
	return service.New(
		cfg.log,
		cfg.root,
		cfg.isPretty,
		fileStorage,
		binaryService.New(fileStorage, cfg.isPretty),
		entryService.New(fileStorage, cfg.isPretty),
		folderService.New(fileStorage, cfg.isPretty),
	)
}
