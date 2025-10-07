//go:build unix

package main

import (
	// for USR1 (stack trace) and USR2 (heap profile) signals.
	// See README for usage.
	_ "github.com/tam7t/sigprof"
)
