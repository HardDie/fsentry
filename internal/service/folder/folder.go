package folder

import (
	"sync"

	"github.com/HardDie/fsentry/internal/entity"
	repositoryEntry "github.com/HardDie/fsentry/internal/repository/entry"
	repositoryFolder "github.com/HardDie/fsentry/internal/repository/folder"
	serviceCommon "github.com/HardDie/fsentry/internal/service/common"
	"github.com/HardDie/fsentry/internal/utils"
	"github.com/HardDie/fsentry/pkg/fsentry_error"
	"github.com/HardDie/fsentry/pkg/fsentry_types"
)

type Folder interface {
	CreateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error)
	GetFolder(name string, path ...string) (*entity.FolderInfo, error)
	MoveFolder(oldName, newName string, path ...string) (*entity.FolderInfo, error)
	UpdateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error)
	RemoveFolder(name string, path ...string) error
	DuplicateFolder(srcName, dstName string, path ...string) (*entity.FolderInfo, error)
	UpdateFolderNameWithoutTimestamp(name, newName string, path ...string) error
}

type folder struct {
	root string
	rwm  *sync.RWMutex

	isPretty bool

	repFolder repositoryFolder.Folder
	repEntry  repositoryEntry.Entry
	common    serviceCommon.Common
}

func NewFolder(
	root string,
	rwm *sync.RWMutex,
	isPretty bool,
	repFolder repositoryFolder.Folder,
	repEntry repositoryEntry.Entry,
	common serviceCommon.Common,
) Folder {
	return &folder{
		root:      root,
		rwm:       rwm,
		isPretty:  isPretty,
		repFolder: repFolder,
		repEntry:  repEntry,
		common:    common,
	}
}

// CreateFolder you can use this method to create a folder within the repository.
// name - Name of the folder to be created.
// data - If you want to store some payload inside the json metadata you can pass it here.
// path - Optional value if you want to create a folder inside an existing folder.
// If you want to create a folder in the root of the storage, you can leave this value empty.
//
// As the result you receive JSON metadata file that was created in the created folder.
//
// Examples:
//
// Create a folder in the root of the storage:
//
//	resp, err := db.CreateFolder("f1", nil)
//	if err != nil {
//		panic(err)
//	}
//
// Create a folder inside an existing folder:
//
//	resp, err := db.CreateFolder("f2", nil, "f1")
//	if err != nil {
//		panic(err)
//	}
func (s *folder) CreateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	id := utils.NameToID(name)
	if id == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsFolderNotExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Create folder
	err = s.repFolder.CreateFolder(fullPath)
	if err != nil {
		return nil, err
	}

	// Create info file
	info := entity.NewFolderInfo(id, name, data, s.isPretty)
	err = s.repFolder.CreateInfo(fullPath, info, s.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s *folder) GetFolder(name string, path ...string) (*entity.FolderInfo, error) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsFolderExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := s.repFolder.GetInfo(fullPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s *folder) MoveFolder(oldName, newName string, path ...string) (*entity.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(oldName) == "" || utils.NameToID(newName) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if source folder exist
	fullOldPath, err := s.common.IsFolderExist(oldName, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := s.repFolder.GetInfo(fullOldPath)
	if err != nil {
		return nil, err
	}

	info.SetName(newName).UpdatedNow()

	// Update info file
	err = s.repFolder.UpdateInfo(fullOldPath, info, s.isPretty)
	if err != nil {
		return nil, err
	}

	// If folders have same ID
	if utils.NameToID(oldName) == utils.NameToID(newName) {
		return info, nil
	}

	// Check if destination folder not exist
	fullNewPath, err := s.common.IsFolderNotExist(newName, path...)
	if err != nil {
		return nil, err
	}

	// Rename folder
	err = s.repEntry.MoveObject(fullOldPath, fullNewPath)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s *folder) UpdateFolder(name string, data interface{}, path ...string) (*entity.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsFolderExist(name, path...)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := s.repFolder.GetInfo(fullPath)
	if err != nil {
		return nil, err
	}

	err = info.UpdateData(data, s.isPretty)
	if err != nil {
		return nil, err
	}

	// Update info file
	err = s.repFolder.UpdateInfo(fullPath, info, s.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s *folder) RemoveFolder(name string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsFolderExist(name, path...)
	if err != nil {
		return err
	}

	err = s.repFolder.RemoveFolder(fullPath)
	if err != nil {
		return err
	}

	return nil
}
func (s *folder) DuplicateFolder(srcName, dstName string, path ...string) (*entity.FolderInfo, error) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(srcName) == "" || utils.NameToID(dstName) == "" {
		return nil, fsentry_error.ErrorBadName
	}

	// Check if source folder exist
	fullSrcPath, err := s.common.IsFolderExist(srcName, path...)
	if err != nil {
		return nil, err
	}

	// Check if destination folder not exist
	fullDstPath, err := s.common.IsFolderNotExist(dstName, path...)
	if err != nil {
		return nil, err
	}

	// Copy folder
	err = s.repFolder.CopyFolder(fullSrcPath, fullDstPath)
	if err != nil {
		return nil, err
	}

	// Get info from file
	info, err := s.repFolder.GetInfo(fullDstPath)
	if err != nil {
		return nil, err
	}

	info.SetName(dstName).FlushTime()

	// Update info file
	err = s.repFolder.CreateInfo(fullDstPath, info, s.isPretty)
	if err != nil {
		return nil, err
	}

	return info, nil
}
func (s *folder) UpdateFolderNameWithoutTimestamp(name, newName string, path ...string) error {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	if utils.NameToID(name) == "" || utils.NameToID(newName) == "" {
		return fsentry_error.ErrorBadName
	}

	fullPath, err := s.common.IsFolderExist(name, path...)
	if err != nil {
		return err
	}

	// Get info from file
	info, err := s.repFolder.GetInfo(fullPath)
	if err != nil {
		return err
	}

	info.Id = utils.NameToID(newName)
	info.Name = fsentry_types.QuotedString(newName)

	// Update info file
	err = s.repFolder.CreateInfo(fullPath, info, s.isPretty)
	if err != nil {
		return err
	}

	return nil
}
