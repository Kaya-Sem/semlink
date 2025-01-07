package cmd

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/sys/unix"
)

// set the value of the key (semlinkXattrKey) to (value)
func setXattr(path string, semlinkXattrKey string, value string) {

	err := unix.Setxattr(path, semlinkXattrKey, []byte(value), 0)
	if err != nil {
		log.Fatalf("Failed to set xattr: %v", err)
	}
}

func getXattr(path string, semlinkXattrKey string) (string, error) {
	value := make([]byte, 1024)
	vLen, err := unix.Getxattr(path, semlinkXattrKey, value)
	if err != nil {
		if err == unix.ENODATA {
			// Key not found
			fmt.Printf("Nothing found for %s with key %s\n", path, semlinkXattrKey)
			return "", nil
		}

		return "", fmt.Errorf("failed to get xattr value for %s: %w", path, err)
	}
	rawValue := string(value[:vLen])
	return rawValue, nil
}

func getSemlinkType(path string) string {
	folderType, err := getXattr(path, semlinkTypeXattrKey)

	if err != nil {
		logFatalWithCaller("error?", err)
	}

	return folderType
}

func getSemlinkTags(path string) []string {
	tagString, err := getXattr(path, semlinkTagXattrKey)

	if err != nil {
		logFatalWithCaller("couldnt get tags", err)
	}

	tags := parseTags(tagString)

	return tags
}

func parseTags(tagString string) []string {
	if tagString == "" {
		return []string{}
	}

	// Split the string by comma and trim spaces
	tags := strings.Split(tagString, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}
