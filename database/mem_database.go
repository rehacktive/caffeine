package database

import (
	"fmt"
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

func (mb *MemDatabase) Upsert(namespace string, key string, value []byte) *DbError {
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

func (mb *MemDatabase) Get(namespace string, key string) ([]byte, *DbError) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ns, ok := mb.namespaces[namespace]
	if !ok {
		return nil, &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace '%v' does not exist.", namespace),
		}
	}
	val, ok := ns.data[key]
	if !ok {
		return nil, &DbError{
			ErrorCode: ID_NOT_FOUND,
			Message:   fmt.Sprintf("value not found in namespace '%v' for key '%v'", namespace, key),
		}
	}
	return val, nil
}

func (mb *MemDatabase) GetAll(namespace string) (map[string][]byte, *DbError) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ns, ok := mb.namespaces[namespace]
	if !ok {
		return nil, &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace '%v' does not exist.", namespace),
		}
	}
	return ns.data, nil
}

func (mb *MemDatabase) Delete(namespace string, key string) *DbError {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ns, ok := mb.namespaces[namespace]
	if !ok {
		return &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace '%v' does not exist.", namespace),
		}
	}
	_, ok = ns.data[key]
	if !ok {
		return &DbError{
			ErrorCode: ID_NOT_FOUND,
			Message:   fmt.Sprintf("value not found in namespace '%v' for key '%v'", namespace, key),
		}
	}

	delete(ns.data, key)
	return nil
}

func (mb *MemDatabase) DeleteAll(namespace string) *DbError {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	_, ok := mb.namespaces[namespace]
	if !ok {
		return &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace '%v' does not exist.", namespace),
		}
	}
	delete(mb.namespaces, namespace)
	return nil
}

func (mb *MemDatabase) GetNamespaces() []string {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ret := make([]string, 0)
	for k := range mb.namespaces {
		ret = append(ret, k)
	}
	return ret
}
