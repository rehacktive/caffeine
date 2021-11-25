package database

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type StorageDatabase struct {
	RootDirPath string
}

func (s *StorageDatabase) Init() {
	err := os.MkdirAll(s.RootDirPath, os.ModePerm)
	if err != nil {
		log.Fatal("error on StorageDatabase Init: ", err)
	}
}

func (s *StorageDatabase) Upsert(namespace string, key string, value []byte) *DbError {
	err := s.ensureNamespace(namespace)
	if err != nil {
		return &DbError {
			ErrorCode: FILESYSTEM_ERROR,
			Message: fmt.Sprintf("%v", err),
		}
	}
	filePath := s.getFilePath(namespace, key)

	_, statErr := os.Stat(filePath)
	if statErr == nil || errors.Is(statErr, os.ErrNotExist) {
		err = os.WriteFile(filePath, value, os.ModePerm)
		if err != nil {
			return &DbError {
				ErrorCode: FILESYSTEM_ERROR,
				Message: fmt.Sprintf("%v", err),
			}
		}
	}
	return nil
}

func (s *StorageDatabase) Get(namespace string, key string) ([]byte, *DbError) {
	filePath := s.getFilePath(namespace, key)
	bytes, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, &DbError {
			ErrorCode: FILESYSTEM_ERROR,
			Message: fmt.Sprintf("%v", err),
		}
	} else {
		return bytes, nil
	}
}

func (s *StorageDatabase) GetAll(namespace string) (map[string][]byte, *DbError) {
	result := make(map[string][]byte)

	docs, readDirErr := ioutil.ReadDir(s.getNamespacePath(namespace))
	if readDirErr != nil {
		return nil, &DbError {
			ErrorCode: FILESYSTEM_ERROR,
			Message: fmt.Sprintf("%v", readDirErr),
		}
	}
	for _, doc := range docs {
		keyParts := strings.SplitN(doc.Name(), ".", 2)
		if len(keyParts) != 2 || keyParts[1] != "json" {
			continue
		}
		rawKey := keyParts[0]
		var err *DbError
		result[rawKey], err = s.Get(namespace, rawKey)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *StorageDatabase) Delete(namespace string, key string) *DbError {
	filePath := s.getFilePath(namespace, key)

	_, err := os.Stat(filePath)
	if err != nil {
		return &DbError {
			ErrorCode: ID_NOT_FOUND,
			Message:   fmt.Sprintf("value not found in namespace '%v' for key '%v'", namespace, key),
		}
	}

	err = os.Remove(filePath)

	if err != nil {
		return &DbError {
			ErrorCode: FILESYSTEM_ERROR,
			Message: fmt.Sprintf("%v", err),
		}
	} else {
		return nil
	}
}

func (s *StorageDatabase) DeleteAll(namespace string) *DbError {
	err := os.RemoveAll(s.getNamespacePath(namespace))

	if err != nil {
		return &DbError {
			ErrorCode: FILESYSTEM_ERROR,
			Message: fmt.Sprintf("%v", err),
		}
	} else {
		return nil
	}
}

func (s *StorageDatabase) GetNamespaces() []string {
	results := make([]string, 0)

	namespaces, err := os.ReadDir(s.RootDirPath)
	if err != nil {
		fmt.Println(err)
		return results
	}

	for _, ns := range namespaces {
		if ns.IsDir() {
			results = append(results, ns.Name())
		}
	}
	return results
}

func (s *StorageDatabase) ensureNamespace(namespace string) error {
	path := s.getNamespacePath(namespace)
	return os.MkdirAll(path, os.ModePerm)
}

func (s *StorageDatabase) getFilePath(namespace, key string) string {
	return filepath.Join(s.getNamespacePath(namespace), fmt.Sprintf("%s.json", key))
}

func (s *StorageDatabase) getNamespacePath(namespace string) string {
	return filepath.Join(s.RootDirPath, namespace)
}
