package binary

import (
	"sync"

	repositoryBinary "github.com/HardDie/fsentry/internal/repository/binary"
	repositoryEntry "github.com/HardDie/fsentry/internal/repository/entry"
	serviceCommon "github.com/HardDie/fsentry/internal/service/common"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

type Binary interface {
	CreateBinary(name string, data []byte, path ...string) error
	GetBinary(name string, path ...string) ([]byte, error)
	MoveBinary(oldName, newName string, path ...string) error
	UpdateBinary(name string, data []byte, path ...string) error
	RemoveBinary(name string, path ...string) error
}

type binary struct {
	root string
	rwm  *sync.RWMutex

	isPretty bool

	repEntry  repositoryEntry.Entry
	repBinary repositoryBinary.Binary
	common    serviceCommon.Common
}

func NewBinary(
	root string,
	rwm *sync.RWMutex,
	isPretty bool,
	repEntry repositoryEntry.Entry,
	repBinary repositoryBinary.Binary,
	common serviceCommon.Common,
) Binary {
	return &binary{
		root:      root,
		rwm:       rwm,
		isPretty:  isPretty,
		repEntry:  repEntry,
		repBinary: repBinary,
		common:    common,
	}
}

func (s *binary) CreateBinary(name string, data []byte, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsBinaryNotExist(name, path...)
	if err != nil {
		return err
	}

	err = s.repBinary.CreateBinary(fullPath, data)
	if err != nil {
		return err
	}

	return nil
}
func (s *binary) GetBinary(name string, path ...string) ([]byte, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsBinaryExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get data from file
	data, err := s.repBinary.GetBinary(fullPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func (s *binary) MoveBinary(oldName, newName string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	// Check if source binary exist
	fullOldPath, err := s.common.IsBinaryExist(oldName, path...)
	if err != nil {
		return err
	}

	// Check if destination binary not exist
	fullNewPath, err := s.common.IsBinaryNotExist(newName, path...)
	if err != nil {
		return err
	}

	// Rename binary
	err = s.repEntry.MoveObject(fullOldPath, fullNewPath)
	if err != nil {
		return err
	}

	return nil
}
func (s *binary) UpdateBinary(name string, data []byte, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsBinaryExist(name, path...)
	if err != nil {
		return err
	}

	// Update binary file
	err = s.repBinary.UpdateBinary(fullPath, data)
	if err != nil {
		return err
	}

	return nil
}
func (s *binary) RemoveBinary(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsBinaryExist(name, path...)
	if err != nil {
		return err
	}

	err = s.repBinary.RemoveBinary(fullPath)
	if err != nil {
		return err
	}

	return nil
}
