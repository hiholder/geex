package util

import "os"

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
