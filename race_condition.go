package main

import (
	"fmt"
	"sync"
)

func processData(wg *sync.WaitGroup, result []int, index int, data int) {
	// wg: shared WaitGroup to track goroutines
	// result: slice (not a pointer this time)
	// index: unique position this goroutine will write to
	// data: the input value

	defer wg.Done()

	result[index] = data * 2
	
	// Each goroutine writes to a UNIQUE index in the slice
	// No other goroutine writes to this same index
	//
	// WHY THIS FIXES IT:
	// - No shared mutation of slice structure (no append)
	// - No resizing or reallocation
	// - Each goroutine operates on its own memory location	
}

func main() {
	var wg sync.WaitGroup
	// Tracks how many goroutines are running

	input := []int{1, 2, 3, 4, 5}
	// Input slice

	result := make([]int, len(input))
	// PREALLOCATED SLICE
	//
	// Creates: [0, 0, 0, 0, 0]
	// Length is FIXED and matches input
	//
	// WHY THIS FIXES IT:
	// - No need for append()
	// - No resizing during execution
	// - Memory is allocated upfront

	for i, data := range input {
		// i = index (0,1,2,3,4)
		// data = value (1,2,3,4,5)

		wg.Add(1)
		// Increment WaitGroup counter

		go processData(&wg, result, i, data)
		// Start goroutine
		//
		// KEY DIFFERENCE:
		// - Passing index (i)
		// - NOT passing pointer to slice
		//
		// Each goroutine now knows EXACTLY where to write
	}

	wg.Wait()
	// Wait until all goroutines finish

	fmt.Println(result)
	// ALWAYS outputs: [2 4 6 8 10]
	//
	// WHY:
	// - Even if goroutines finish out of order,
	//   they write to fixed positions
}