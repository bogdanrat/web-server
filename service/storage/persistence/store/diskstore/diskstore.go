package diskstore

import (
	"fmt"
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/storage/lib"
	"github.com/bogdanrat/web-server/service/storage/persistence/store"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DiskStore struct {
	Path string
}

func New(path string) store.Store {
	return &DiskStore{
		Path: path,
	}
}

func (s *DiskStore) Init() error {
	if err := lib.CreateDirectory(s.Path); err != nil {
		return err
	}
	log.Printf("Initialized Disk Storage Engine in %s\n", s.Path)
	return nil
}

func (s *DiskStore) Put(fileName string, body io.Reader) error {
	name := filepath.Join(s.Path, fileName)
	dir := filepath.Dir(name)
	// create path
	if err := lib.CreateDirectory(dir); err != nil {
		return err
	}

	// create file
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	// copy content
	_, err = io.Copy(file, body)
	if err != nil {
		return err
	}

	return nil
}

func (s *DiskStore) Get(fileName string, writer io.Writer) error {
	name := filepath.Join(s.Path, fileName)
	if exists, err := lib.FileExists(name); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("file %s not found", fileName)
	}

	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func (s *DiskStore) GetAll() ([]*pb.StorageObject, error) {
	objects := make([]*pb.StorageObject, 0)

	err := filepath.WalkDir(s.Path, func(filePath string, d fs.DirEntry, err error) error {
		// avoid directories and hidden files (base == extension, e.g., .DS_Store)
		if !d.IsDir() && path.Base(filePath) != filepath.Ext(filePath) {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			fileInfo, err := file.Stat()
			if err != nil {
				return err
			}

			// remove data/ from the file name
			// s.Path is ./data so we take the base of data, which is 'data'
			fileName := strings.Replace(filePath, path.Base(s.Path)+"/", "", 1)
			object := &pb.StorageObject{
				Key:          fileName,
				Size:         uint64(fileInfo.Size()),
				LastModified: fileInfo.ModTime().Format(time.RFC3339),
			}
			objects = append(objects, object)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return objects, nil
}
func (s *DiskStore) Delete(fileName string) error {
	name := filepath.Join(s.Path, fileName)
	if err := lib.TryRemoveFile(name); err != nil {
		return err
	}
	return nil
}
func (s *DiskStore) DeleteAll(prefix ...string) error {
	// if no prefix was supplied, we are going to delete every file in the storage path (i.d., ./data)
	filesPath := s.Path
	if len(prefix) == 1 {
		filesPath = filepath.Join(s.Path, prefix[0])
	}

	dir, err := os.Open(filesPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// get all contents of the directory associated with 'dir'
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	// remove all contents in dir
	for _, name := range names {
		err := os.RemoveAll(filepath.Join(filesPath, name))
		if err != nil {
			return err
		}
	}

	return nil
}
