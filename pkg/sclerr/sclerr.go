package sclerr

import "io"

// CloseQuietly closes the given object ignoring errors. Useful in defer.
func CloseQuietly(r io.Closer) {
	_ = r.Close()
}
