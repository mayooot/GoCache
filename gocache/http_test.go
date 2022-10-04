package gocache

import (
	"fmt"
	"strings"
	"testing"
)

func TestHasPrefix(t *testing.T) {
	bashPath := "/_gocache/"

	path := "/_gocache/fist/k1"
	path2 := "_gocache/fist/k1"
	path3 := "/_gocache1/fist/k1"
	fmt.Println(strings.HasPrefix(path, bashPath))  // true
	fmt.Println(strings.HasPrefix(path2, bashPath)) // false
	fmt.Println(strings.HasPrefix(path3, bashPath)) // false

}
func TestSplitN(t *testing.T) {
	bashPath := "/_gocache/"

	path := "/_gocache/first/k1"
	parts := strings.SplitN(path[len(bashPath):], "/", 2)
	fmt.Println(parts) // [first k1]

	path4 := "/_gocache/"
	parts4 := strings.SplitN(path4[len(bashPath):], "/", 2)
	fmt.Println(parts4) // []

	path5 := "/_gocache/first"
	parts5 := strings.SplitN(path5[len(bashPath):], "/", 2)
	fmt.Println(parts5) // [first]

	path2 := "/_gocache/first/k1/"
	parts2 := strings.SplitN(path2[len(bashPath):], "/", 2)
	fmt.Println(parts2) // [first k1/]

	path3 := "/_gocache/first/k1/v1"
	parts3 := strings.SplitN(path3[len(bashPath):], "/", 2)
	fmt.Println(parts3) // [first k1/v1]
}
