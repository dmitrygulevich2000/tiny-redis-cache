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
	data := New(0)
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
	data := New(0)
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
	data := New(0)
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

type KeysTestCase struct {
	Pattern string
	Keys []string
	Ttls []time.Duration
	ExpectedResult []string
}

func TestKeys(t *testing.T) {
	tests := []KeysTestCase {
		KeysTestCase {
			Pattern: "h*llo",
			Keys: []string{"hello", "hllo", "hxxxllo", "llo", "hlo"},
			Ttls: []time.Duration{defaultTTL, 2*defaultSleep, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"hllo", "hxxxllo"},  // i also add ttl for "hello" key
		},
		KeysTestCase {
			Pattern: "h[ae]llo",
			Keys: []string{"hello", "hallo", "hxllo", "hllo", "allo"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"hallo", "hello"},
		},
		KeysTestCase {
			Pattern: "h[^ae]llo",
			Keys: []string{"hello", "hallo", "hxllo", "hllo", "allo"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"hxllo"},
		},
		KeysTestCase {
			Pattern: "h[a-c]llo",
			Keys: []string{"hallo", "hbllo", "hcllo", "hllo", "hhllo", "allo"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"hallo", "hbllo", "hcllo"},
		},
		KeysTestCase {
			Pattern: "h[^a-c]llo",
			Keys: []string{"hallo", "hbllo", "hcllo", "hllo", "hhllo", "allo"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"hhllo"},
		},
		KeysTestCase {
			Pattern: `h[\^\]\\]llo`,
			Keys: []string{"h\\llo", "h^llo", "h]llo", "hllo", "hhllo", "allo"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"h\\llo", "h]llo", "h^llo"},
		},
		KeysTestCase {
			Pattern: "example.com/*",
			Keys: []string{"example.com/", "exampleacom/", "example.com/user", "example.com////"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"example.com/", "example.com////", "example.com/user"},
		},
		KeysTestCase {
			Pattern: "K?",
			Keys: []string{"K", "KK", "KKK"},
			Ttls: []time.Duration{zeroDuration, zeroDuration, zeroDuration},
			ExpectedResult: []string{"KK"},
		},
	}
	
	for num, c:= range tests {
		data := New(0)

		for i, key := range c.Keys {
			data.Set(key, "val", c.Ttls[i])
		}
		time.Sleep(defaultSleep)

		result, err := data.Keys(c.Pattern)
		if err != nil {
			t.Errorf("TestCase %d: Keys: unexpected error %s", num, err.Error())
			break
		}
		sort.Slice(result, func (i,j int) bool {return result[i] < result[j]})

		success := true
		if len(result) != len(c.ExpectedResult) {
			success = false
		} else {
			for i, _ := range result {
				if result[i] != c.ExpectedResult[i] {
					success = false
				}
			}
		}

		if !success {
			t.Errorf("TestCase %d: Keys: expected %v\ngot %v", num, c.ExpectedResult, result)
		}

		data.Close()
	}
}

func TestActiveExpiration(t *testing.T) {
	idata := New(0)
	data := idata.(*kvStorage)
	defer data.Close()

	data.Set("key", "val", defaultTTL)
	time.Sleep(2 * data.resolution)

	data.mutex.RLock()
	size := len(data.data)
	data.mutex.RUnlock()
	if size == 1 {
		t.Fatalf("Expected zero size of the underlying map, got %d\n", size)
	}
}

func TestConcurrentAccess(t *testing.T) {
	kKeys := 10
	kIters := 100
	wg := &sync.WaitGroup{}

	data := New(0)
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
		data := New(0)
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