package fs

import (
	"io"
	"os"
)

// Read reads from the current offset until EOF into buf.
// If buf is too small, it is grown with append. A sized buf with cap >= size
// is the zero-extra-alloc path. io.EOF is success, not an error.
func Read(file *os.File, buf []byte) ([]byte, error) {
	out := buf[:0]
	for {
		if cap(out) == len(out) {
			n := 64
			if c := cap(out); c > 0 {
				n = c * 2
			}
			grown := make([]byte, len(out), n)
			copy(grown, out)
			out = grown
		}
		n, err := file.Read(out[len(out):cap(out)])
		out = out[:len(out)+n]
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, mapError(err)
		}
	}
}
