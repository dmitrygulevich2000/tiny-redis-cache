package storage

import (
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	zeroDuration time.Duration
	defaultTTL = 500*time.Microsecond
	defaultSleep = 2*defaultTTL
)

func TestNoTTL(t *testing.T) {
	data := New()
	defer data.Close()

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
	defer data.Close()

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
	defer data.Close()

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

func TestKeys(t *testing.T) {
	pattern := "h*llo"
	keys := []string{"hello", "hllo", "hxxxllo", "llo", "hlo"}
	ttls := []time.Duration{defaultTTL, 2*defaultSleep, zeroDuration, zeroDuration, zeroDuration}
	expectedResult := []string{"hllo", "hxxxllo"}  // i also add ttl for "hello" key

	data := New()
	defer data.Close()

	for i, key := range keys {
		data.Set(key, "val", ttls[i])
	}
	time.Sleep(defaultSleep)

	result, err := data.Keys(pattern)
	if err != nil {
		t.Fatalf("Keys: unexpected error %s", err.Error())
	}
	sort.Slice(result, func (i,j int) bool {return result[i] < result[j]})

	success := true
	if len(result) != len(expectedResult) {
		success = false
	} else {
		for i, _ := range result {
			if result[i] != expectedResult[i] {
				success = false
			}
		}
	}

	if !success {
		t.Fatalf("Keys: expected %v\ngot %v", expectedResult, result)
	}
}

func TestActiveExpiration(t *testing.T) {
	data := New()
	defer data.Close()

	data.Set("key", "val", defaultTTL)
	time.Sleep(2 * data.resolution)

	size := len(data.data)
	if size == 1 {
		t.Fatalf("Expected zero size of the underlying map, got %d\n", size)
	}
}

func TestConcurrentAccess(t *testing.T) {
	kKeys := 10
	kIters := 100
	wg := &sync.WaitGroup{}

	data := New()
	defer data.Close()

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


func BenchmarkConcurrentAccess(b *testing.B) {
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
		data.Close()
    }
}