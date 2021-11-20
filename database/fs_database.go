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

func (s *StorageDatabase) Upsert(namespace string, key string, value []byte) error {
	err := s.ensureNamespace(namespace)
	if err != nil {
		return err
	}
	filePath := s.getFilePath(namespace, key)

	_, err = os.Stat(filePath)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(filePath, value, os.ModePerm)
		return err
	}
	return err
}

func (s *StorageDatabase) Get(namespace string, key string) ([]byte, error) {
	filePath := s.getFilePath(namespace, key)
	return ioutil.ReadFile(filepath.Clean(filePath))
}

func (s *StorageDatabase) GetAll(namespace string) (map[string][]byte, error) {
	result := make(map[string][]byte)

	docs, err := ioutil.ReadDir(s.getNamespacePath(namespace))
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		keyParts := strings.SplitN(doc.Name(), ".", 2)
		if len(keyParts) != 2 || keyParts[1] != "json" {
			continue
		}
		rawKey := keyParts[0]
		result[rawKey], err = s.Get(namespace, rawKey)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *StorageDatabase) Delete(namespace string, key string) error {
	filePath := s.getFilePath(namespace, key)

	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return os.Remove(filePath)
}

func (s *StorageDatabase) DeleteAll(namespace string) error {
	return os.RemoveAll(s.getNamespacePath(namespace))
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
