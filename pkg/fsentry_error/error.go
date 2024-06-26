package fsentry_error

import (
	"errors"
	"fmt"
)

var (
	ErrorBadName         = fmt.Errorf("bad name")
	ErrorBadPath         = fmt.Errorf("bad path")
	ErrorExist           = fmt.Errorf("object exist")
	ErrorNotExist        = fmt.Errorf("object not exist")
	ErrorPermissions     = fmt.Errorf("not enough permissions")
	ErrorNotFile         = fmt.Errorf("not file")
	ErrorNotDirectory    = fmt.Errorf("not directory")
	ErrorInternal        = fmt.Errorf("internal error")
	ErrorFolderCorrupted = fmt.Errorf("foler corrupted")
	// windows.
	ErrorIncorrectFunction = fmt.Errorf("incorrect function")
	ErrorIsDirectory       = fmt.Errorf("is directory")
)

func Wrap(err, localErr error) error {
	return errors.Join(err, localErr)
}
