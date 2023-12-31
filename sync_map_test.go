package syncmap

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {
	m := New(map[string]int{})

	var wg sync.WaitGroup
	var N int = 10000

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Insert(fmt.Sprint(n), n, false)
		}(i)
	}

	wg.Wait()

	if m.Len() != N {
		t.Fatalf("mismatched count: expected %v, got %v\n", N, m.Len())
	}
}

func TestSyncMapGet(t *testing.T) {
	m := New(map[string]int{"a": 1})

	type test struct {
		k           string
		v           int
		e           error
		expectError bool
	}

	tests := []test{
		{
			k:           "a",
			v:           1,
			e:           nil,
			expectError: false,
		},
		{
			k:           "c",
			v:           2,
			e:           ErrMissingKey,
			expectError: true,
		},
	}

	for _, test := range tests {

		v, err := m.Get(test.k)
		if test.expectError {
			if err == nil {
				t.Fatal("expected error but none returned")
			}
			if !errors.Is(err, test.e) {
				t.Fatalf("unexpected error (%v) returned", err)
			}
		} else {
			if err != nil {
				t.Fatalf("unexpected error (%v) returned, when no error should be returned", err)
			} else {
				if v != test.v {
					t.Fatalf("unexpected value (%v) returned; expected (%v)", v, test.v)
				}
			}
		}
	}
}

func TestSyncMapContains(t *testing.T) {
	m := New(map[string]int{"a": 1})

	type test struct {
		k     string
		found bool
	}

	tests := []test{
		{
			k:     "a",
			found: true,
		},
		{
			k:     "b",
			found: false,
		},
	}

	for _, test := range tests {

		v := m.Contains(test.k)
		if v && !test.found {
			t.Fatal("returned found when expected not to find")
		} else if !v && test.found {
			t.Fatal("returned not found when expected to find")
		}
	}
}

func TestSyncMapGetKeys(t *testing.T) {
	m := New(map[string]int{"c": 1, "b": 2, "a": 3})
	keys := m.GetKeys()
	if strings.Join(keys, "||") != strings.Join([]string{"a", "b", "c"}, "||") {
		t.Fatalf("unexpected keys returned (%v)", keys)
	}
}

func TestSyncMapDelete(t *testing.T) {
	m := New(map[string]int{"c": 1, "b": 2, "a": 3})

	m.Remove("a")
	m.Remove("c")
	m.Remove("aa")

	if fmt.Sprint(m) != "map[b:2]" {
		t.Fatalf("unexpected post deletion state (%v)", m)
	}
}

func TestSyncMapBytes(t *testing.T) {

	c1 := New(map[string]int{"a": 1, "b": -1})
	b, err := c1.Bytes()
	if err != nil {
		t.Fatalf("unexpected error during Bytes(): %v", err)
	}

	c2 := New(map[string]int{})

	err = c2.Merge(b)
	if err != nil {
		t.Fatalf("unexpected error during Merge(): %v", err)
	}

	if c1.String() != c2.String() {
		t.Fatalf("mismatch: expected %v, got %v", c1, c2)
	}
}

func TestSyncMapBytes2(t *testing.T) {

	c1 := New(map[string]int{"a": 1, "b": -1})
	b, err := c1.Bytes()
	if err != nil {
		t.Fatalf("unexpected error during Bytes(): %v", err)
	}

	c2 := New(map[string]int{"a": 2})

	err = c2.Merge(b)
	if err != nil {
		t.Fatalf("unexpected error during Merge(): %v", err)
	}

	expected := "map[a:2 b:-1]"
	if c2.String() != expected {
		t.Fatalf("mismatch: expected %q, got %q", expected, c2)
	}
}

func TestSyncMapBytes3(t *testing.T) {

	c1 := New(map[string]int{"a": 1, "b": -1})
	b, err := c1.Bytes()
	if err != nil {
		t.Fatalf("unexpected error during Bytes(): %v", err)
	}

	c2 := New(map[string]int{"c": -3})

	err = c2.Merge(b)
	if err != nil {
		t.Fatalf("unexpected error during Merge(): %v", err)
	}

	expected := "map[a:1 b:-1 c:-3]"
	if c2.String() != expected {
		t.Fatalf("mismatch: expected %q, got %q", expected, c2)
	}
}

func ExampleNew() {
	c := New(map[string]int{"x": 0, "y": 0})

	// Adds z
	c.Insert("a", 1, false)

	// Updates z without raising an error
	c.Insert("a", 2, false)

	c.Remove("y")

	fmt.Println(c)
	// Output: map[a:2 x:0]
}
