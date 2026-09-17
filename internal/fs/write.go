package fs

import "os"

// Write writes all of data to file. It does not create, seek, sync, or close.
func Write(file *os.File, data []byte) error {
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			return mapError(err)
		}
		data = data[n:]
	}
	return nil
}
