package lib

import (
	"os"
	"path/filepath"
)

var (
	osStat       = os.Stat
	osIsNotExist = os.IsNotExist
)

// FileExists return flag whether a given file exists
// and operation error if an unclassified failure occurs.
func FileExists(path string) (bool, error) {
	_, err := osStat(path)
	if err != nil {
		if osIsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateDirectory makes directory with a given name
// making all parent directories if necessary.
func CreateDirectory(dirPath string) error {
	var dPath string
	var err error
	if !filepath.IsAbs(dirPath) {
		dPath, err = filepath.Abs(dirPath)
		if err != nil {
			return err
		}
	} else {
		dPath = dirPath
	}
	exists, err := FileExists(dPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return os.MkdirAll(dPath, os.ModePerm)
}

// TryRemoveFile gives a try removing the file
// only ignoring an error when the file does not exist.
func TryRemoveFile(filePath string) (err error) {
	err = os.RemoveAll(filePath)
	if os.IsNotExist(err) {
		err = nil
		return
	}
	return
}
