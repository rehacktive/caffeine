package database

import (
	"errors"
	"sync"
)

type MemDatabase struct {
	mu         sync.Mutex
	namespaces map[string]namespace
}

type namespace struct {
	data map[string][]byte
}

func newNamespace() namespace {
	return namespace{
		data: make(map[string][]byte),
	}
}

func (mb *MemDatabase) Init() {
	mb.namespaces = make(map[string]namespace)
}

func (mb *MemDatabase) Upsert(namespace string, key string, value []byte) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ns, ok := mb.namespaces[namespace]
	if !ok {
		ns = newNamespace()
		mb.namespaces[namespace] = ns
	}
	ns.data[key] = value
	return nil
}

func (mb *MemDatabase) Get(namespace string, key string) ([]byte, error) {
	ns, ok := mb.namespaces[namespace]
	if !ok {
		return nil, errors.New("not found")
	}
	val, ok := ns.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

func (mb *MemDatabase) GetAll(namespace string) (map[string][]byte, error) {
	ns, ok := mb.namespaces[namespace]
	if !ok {
		return nil, errors.New("not found")
	}
	return ns.data, nil
}

func (mb *MemDatabase) Delete(namespace string, key string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ns, ok := mb.namespaces[namespace]
	if !ok {
		return errors.New("not found")
	}
	delete(ns.data, key)
	return nil
}

func (mb *MemDatabase) DeleteAll(namespace string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	_, ok := mb.namespaces[namespace]
	if !ok {
		return errors.New("not found")
	}
	delete(mb.namespaces, namespace)
	return nil
}

func (mb *MemDatabase) GetNamespaces() []string {
	ret := []string{}
	for k := range mb.namespaces {
		ret = append(ret, k)
	}
	return ret
}
