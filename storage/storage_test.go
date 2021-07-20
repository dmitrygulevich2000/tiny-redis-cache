package storage

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	zeroDuration time.Duration
	defaultTTL = 500*time.Microsecond
	defaultSleep = time.Millisecond
)

func TestNoTTL(t *testing.T) {
	data := New()

	data.Set("key", "val1", zeroDuration)
	val, _ := data.Get("key")
	if val != "val1" {
		t.Fatalf("Subtest 1: Get key1: expected %s, got %s\n", "val1", val)
	}

	data.Set("key", "val2", zeroDuration)
	val, _ = data.Get("key")
	if val != "val2" {
		t.Fatalf("Subtest 2: Get key: expected %s, got %s\n", "val2", val)
	}

	deleted := data.Delete("key")
	if deleted != 1 {
		t.Fatalf("Subtest 3: Delete key: expected %d, got %d\n", 1, deleted)
	}
	val, exists := data.Get("key")
	if exists {
		t.Fatalf("Subtest 4: Get key: expected nothing, got %s\n", val)
	}

	deleted = data.Delete("key")
	if deleted != 0 {
		t.Fatalf("Subtest 5: Delete key: expected %d, got %d\n", 0, deleted)
	}
}

func TestWithTTL(t *testing.T) {
	data := New()

	data.Set("key", "val", defaultTTL)
	val, _ := data.Get("key")
	if val != "val" {
		t.Fatalf("Subtest 1: Get key: expected %s, got %s", "val\n", val)
	}
	time.Sleep(defaultSleep)
	val, exists := data.Get("key")
	if exists {
		t.Fatalf("Subtest 1: Get key: expected nothing, got %s\n", val)
	}
	deleted := data.Delete("key")
	if deleted != 0 {
		t.Fatalf("Subtest 1: Delete key: expected %d, got %d\n", 0, deleted)
	}

	data.Set("key", "val1", defaultTTL)
	data.Set("key", "val2", zeroDuration)
	time.Sleep(defaultSleep)
	val, _ = data.Get("key")
	if val != "val2" {
		t.Fatalf("Subtest 2: Get key: expected %s, got %s", "val2\n", val)
	}
}

func TestDeleteManyKeys(t *testing.T) {
	data := New()

	data.Set("key1", "val", zeroDuration)
	data.Set("key2", "val", zeroDuration)
	deleted := data.Delete("key1", "key2")
	if deleted != 2 {
		t.Fatalf("Subtest 1: Delete key1, key2: expected %d, got %d\n", 2, deleted)
	}

	data.Set("key1", "val", zeroDuration)
	data.Set("key2", "val", zeroDuration)
	deleted = data.Delete("key1")
	if deleted != 1 {
		t.Fatalf("Subtest 2: Delete key1: expected %d, got %d\n", 1, deleted)
	}
	deleted = data.Delete("key1", "key2")
	if deleted != 1 {
		t.Fatalf("Subtest 2: Delete key1, key2: expected %d, got %d\n", 1, deleted)
	}

	data.Set("key1", "val", defaultTTL)
	data.Set("key2", "val", zeroDuration)
	time.Sleep(defaultSleep)
	deleted = data.Delete("key1", "key2")
	if deleted != 1 {
		t.Fatalf("Subtest 3: Delete key1, key2: expected %d, got %d\n", 1, deleted)
	}
}

func TestConcurrent(t *testing.T) {
	kKeys := 10
	kIters := 100
	wg := &sync.WaitGroup{}

	data := New()

	for i := 0; i < kKeys; i += 1 {
		key := "key" + strconv.Itoa(i)
		data.Set(key, 0, zeroDuration)
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			for j := 0; j < kIters; j += 1 {
				ival, _ := data.Get(key)
				runtime.Gosched()
				val := ival.(int)
				val += 1
				data.Set(key, val, zeroDuration)
				runtime.Gosched()
			}
			wg.Done()
		}(wg)
	}

	wg.Wait()

	for i := 0; i < kKeys; i += 1 {
		key := "key" + strconv.Itoa(i)
		ival, _ := data.Get(key)
		val := ival.(int)
		if val != kIters{
			t.Fatalf("Get %s: expected %d, got %d\n", key, kIters, val)
		}
	}
}


func BenchmarkConcurrent(b *testing.B) {
	kKeys := 10
	kIters := 100

	for i := 0; i < b.N; i++ {
		data := New()
		wg := &sync.WaitGroup{}
		
		for i := 0; i < kKeys; i += 1 {
			key := "key" + strconv.Itoa(i)
			data.Set(key, 0, zeroDuration)
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < kIters; j += 1 {
					ival, _ := data.Get(key)
					val := ival.(int)
					val += 1
					data.Set(key, val, zeroDuration)
				}
				wg.Done()
			}(wg)
			
		}
	
		wg.Wait()
    }
}