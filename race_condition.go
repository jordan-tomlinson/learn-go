package main // Defines the main package (entry point of the program)

import (
	"fmt"  // Used for printing output to the console
	"sync" // Provides synchronization primitives (like WaitGroup)
)

func processData(wg *sync.WaitGroup, result *[]int, data int) {
	// wg: pointer to a WaitGroup (shared across goroutines)
	// result: pointer to a slice (shared mutable state ⚠️)
	// data: the input value to process

	defer wg.Done() // Signals that this goroutine is done when the function exits

	processData := data * 2
	// Creates a local variable (unfortunately named same as function)
	// Doubles the input value (this part is fine)

	*result = append(*result, processData)
	// ❌ PROBLEM LINE
	// Dereferences the slice pointer and appends a value to it
	//
	// WHY THIS BREAKS:
	// 1. append() is NOT thread-safe
	// 2. Multiple goroutines call this at the same time
	// 3. append() may:
	//    - Modify the slice length
	//    - Reallocate the underlying array
	//    - Move data in memory
	//
	// This causes:
	// - Data races (multiple writes at once)
	// - Lost updates (some values overwritten or skipped)
	// - Corrupted slice state in worst case
}

func main() {
	var wg sync.WaitGroup
	// WaitGroup tracks how many goroutines are running

	input := []int{1, 2, 3, 4, 5}
	// Input slice of numbers

	result := []int{}
	// Shared slice where results will be stored
	// ⚠️ Starts empty and will grow via append (unsafe with concurrency)

	for _, data := range input {
		// Loop through each value in the input slice

		wg.Add(1)
		// Increment WaitGroup counter (we’re about to start a goroutine)

		go processData(&wg, &result, data)
		// ❗ Starts a goroutine (runs concurrently)
		//
		// WHAT THIS MEANS:
		// - Multiple processData functions run at the same time
		// - They ALL share the same "result" slice
		// - They ALL try to append at the same time → 💥 race condition
	}

	wg.Wait()
	// Blocks until all goroutines call wg.Done()

	fmt.Println(result)
	// Prints the result slice
	// ❗ Output is NON-DETERMINISTIC due to race condition
	// Examples:
	// - [2 4 6 8 10] (lucky)
	// - [2 6 4 10 8] (out of order)
	// - [2 4 4 8] (missing values)
}