package main

import "fmt"

// closure func signature must match return func signature
func gen() func() int {
	var i int
	return func() int {
		i++
		return i
	}
}

func main() {
	counter := gen()
	for i := 1; i <= 5; i++ {
		fmt.Printf("iteration %d - counter value %d\n", i, counter())
	}
}
