//go:build (!windows && !darwin) || !cgo

package dragout

import "unsafe"

const Available = false

func Start(window unsafe.Pointer, paths []string) (Result, error) {
	return "", ErrUnsupported
}
