package cmd

import (
	"log"

	"golang.org/x/sys/unix"
)

// set the value of the key (semlinkXattrKey) to (value)
func setXattr(path string, semlinkXattrKey string, value string) {

	err := unix.Setxattr(path, semlinkXattrKey, []byte(value), 0)
	if err != nil {
		log.Fatalf("Failed to set xattr: %v", err)
	}
}
