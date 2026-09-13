package profile

import (
	"fmt"
	"io"
	"os"
)

// ReadBoundedFile reads path up to maxBytes (+1 byte to detect overflow).
func ReadBoundedFile(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxProfileBytes
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	limited := io.LimitReader(f, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%w: file exceeds %d bytes", ErrProfileTooLarge, maxBytes)
	}
	return data, nil
}
