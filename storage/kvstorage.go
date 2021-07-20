package storage

import (
	"regexp"
	"sync"
	"time"
)

var (
	initialSize = 16
)


type KVStorage struct {
	data map[string]interface{}
	expires map[string]time.Time
	mutex sync.RWMutex
}

func New() *KVStorage {
	storage := &KVStorage{
		data: make(map[string]interface{}, initialSize),
		expires: make(map[string]time.Time, initialSize),
	}
	return storage
}


func (s *KVStorage) Set(key string, value interface{}, ttl time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
	if ttl > 0 {
		s.expires[key] = time.Now().Add(ttl)
	} else {
		delete(s.expires, key)
	}
}

func (s *KVStorage) Delete(keys ...string) int {
	kDeleted := 0
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for _, key := range keys {
		_, exists := s.data[key]
		if exists {

			// dont consider expired keys
			if expires, exists := s.expires[key]; !exists || time.Now().Before(expires) {
				kDeleted += 1
			}
			
			delete(s.data, key)
			delete(s.expires, key)
		}
	}

	return kDeleted
}

func (s *KVStorage) cleanup(key string) (deleted bool) {
	deleted = false
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expires, exists := s.expires[key]
	// must deny write ops after check but before actual deletion
	if exists && time.Now().After(expires) {
		delete(s.data, key)
		delete(s.expires, key)
		deleted = true
	}

	return
}

func (s *KVStorage) Get(key string) (interface{}, bool) {
	s.mutex.RLock()

	value, exists := s.data[key]
	if !exists {
		s.mutex.RUnlock()
		return nil, false
	}
	expires, exists := s.expires[key]
	
	s.mutex.RUnlock()

	if exists && time.Now().After(expires) {
		go s.cleanup(key)
		return nil, false
	}

	return value, true
}

func (s *KVStorage) Keys(pattern string) ([]string, error) {
	expr, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	expr.Longest()
	result := make([]string, 0)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for key, _ := range s.data {
		expires, exists := s.expires[key]
		
		if !exists || time.Now().Before(expires) {
			
			if match := expr.FindString(key); len(match) == len(key) {
				result = append(result, key)
			} // else skip
		} else {
			// TODO: delete key (?)
		}
	}

	return result, nil
}