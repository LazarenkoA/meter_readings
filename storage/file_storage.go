package storage

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"sync"
)

type FileStorage struct {
	mx sync.Map
}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (s *FileStorage) StoreObject(name string, object any) error {
	s.lock(name)
	defer s.unlock(name)

	data, err := yaml.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "yaml marshal error")
	}

	f, err := os.Create(name + ".yaml")
	if err != nil {
		return errors.Wrap(err, "create file error")
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return errors.Wrap(err, "StoreObject error")
	}
	return nil
}

func (s *FileStorage) RestoreObject(name string) (object map[string]interface{}, err error) {
	s.lock(name)
	defer s.unlock(name)

	f, err := os.Open(name + ".yaml")
	if err != nil {
		return nil, errors.Wrap(err, "open file error")
	}

	defer f.Close()

	data, _ := io.ReadAll(f)
	if err := yaml.Unmarshal(data, &object); err != nil {
		return nil, errors.Wrap(err, "yaml unmarshal error")
	}

	return
}

func (s *FileStorage) DeleteObject(name string) error {
	s.lock(name)
	defer s.unlock(name)

	return os.Remove(name + ".yaml")
}

func (s *FileStorage) lock(key string) {
	l, _ := s.mx.LoadOrStore(key, &sync.Mutex{})
	l.(*sync.Mutex).Lock()
}

func (s *FileStorage) unlock(key string) {
	l, _ := s.mx.LoadOrStore(key, &sync.Mutex{})
	l.(*sync.Mutex).Unlock()
}
