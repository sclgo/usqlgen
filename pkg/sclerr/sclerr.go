package sclerr

// CloseQuietly closes the given object ignoring errors. Useful in defer.
func CloseQuietly(r interface{ Close() error }) {
	_ = r.Close()
}
