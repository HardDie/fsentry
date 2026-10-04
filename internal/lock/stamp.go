package lock

import (
	"encoding/binary"
	"os"

	"github.com/HardDie/fsentry/internal/fs"
)

const stampSize = 8

func writeStamp(file *os.File, unixNano int64) error {
	var buf [stampSize]byte
	binary.BigEndian.PutUint64(buf[:], uint64(unixNano))
	if err := fs.WriteAt(file, buf[:], 0); err != nil {
		return err
	}
	if err := fs.Truncate(file, stampSize); err != nil {
		return err
	}
	return fs.Sync(file)
}

// touchStamp writes the stamp for a shared holder: no truncate, no fsync.
// Waiters read it through the page cache. After a crash the OS lock is gone,
// so the stamp does not need to reach the disk.
func touchStamp(file *os.File, unixNano int64) error {
	var buf [stampSize]byte
	binary.BigEndian.PutUint64(buf[:], uint64(unixNano))
	return fs.WriteAt(file, buf[:], 0)
}

func readStamp(file *os.File) (unixNano int64, ok bool, err error) {
	var buf [stampSize]byte
	n, err := fs.ReadAt(file, buf[:], 0)
	if err != nil {
		return 0, false, err
	}
	if n < stampSize {
		return 0, false, nil
	}
	return int64(binary.BigEndian.Uint64(buf[:])), true, nil
}
