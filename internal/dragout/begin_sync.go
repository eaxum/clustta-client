//go:build !darwin || !cgo

package dragout

import "unsafe"

// Begin delivers completion on the UI thread after the synchronous backend returns.
func Begin(window unsafe.Pointer, paths []string, complete func(Result, error)) {
	result, err := Start(window, paths)
	complete(result, err)
}
