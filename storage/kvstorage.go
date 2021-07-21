package storage

import (
	_ "fmt"
	"regexp"
	"sync"
	"time"
)

var (
	initialSize = 16
	defaultResolution = time.Second
)


type KVStorage struct {
	data map[string]interface{}
	expires map[string]time.Time
	
	mutex sync.RWMutex
	done chan struct{}

	resolution time.Duration
}

func New(res time.Duration) *KVStorage {
	storage := &KVStorage{
		data: make(map[string]interface{}, initialSize),
		expires: make(map[string]time.Time, initialSize),
		
		done: make(chan struct{}, 0),
		resolution: defaultResolution,
	}
	if res > 0 {
		storage.resolution = res
	}

	go storage.expirationChecker()
	return storage
}

func (s *KVStorage) Close() {
	close(s.done)

	s.mutex.Lock()
	s.data = nil
	s.expires = nil
	s.mutex.Unlock()
}

func (s *KVStorage) Closed() bool {
	return s.data == nil
}

func (s *KVStorage) Set(key string, value interface{}, ttl time.Duration) {
	if s.Closed() {
		panic("Set over closed storage")
	}

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
	if s.Closed() {
		panic("Delete over closed storage")
	}
	
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
	if s.Closed() {
		panic("Get over closed storage")
	}
	
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

// now works incorrectly
func (s *KVStorage) Keys(pattern string) ([]string, error) {
	if s.Closed() {
		panic("Keys over closed storage")
	}
	
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
		
		if !exists || time.Now().Before(expires) { // not expired
			
			if match := expr.FindString(key); len(match) == len(key) {
				result = append(result, key)
			} // else skip
		} else {  // expired
			// TODO: delete key (?)
		}
	}

	return result, nil
}

func (s *KVStorage) cleanupAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for key, expires := range s.expires {
		if time.Now().After(expires) {
			delete(s.data, key)
			delete(s.expires, key)
		}
	}
}

func (s *KVStorage) expirationChecker() {
	ticker := time.NewTicker(s.resolution)

	for {
		select {
		case <- ticker.C:
			s.cleanupAll()
		case _, ok := <- s.done:
			if !ok {
				ticker.Stop()
				return
			}
		}
	}
}