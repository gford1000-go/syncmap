[![Go Doc](https://pkg.go.dev/badge/github.com/gford1000-go/syncmap.svg)](https://pkg.go.dev/github.com/gford1000-go/syncmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/gford1000-go/syncmap)](https://goreportcard.com/report/github.com/gford1000-go/syncmap)

syncmap
=======

`SynchronisedMap` provides a simple implementation of a concurrency safe map.

The map supports the usual operations, with Insert() operating as both Add and Update.

Maps can be serialised to []byte and then this slice merged into another map instance.  Merge will only insert values for missing keys, so that the receiving map's existing state is not affected. 


```go
func main() {
	c := New(map[string]int{"x": 0, "y": 0})

	c.Insert("z", 1, false)

	c.Remove("y")

	fmt.Println(c)
	// Output: map[x:0 z:1]
}

```
