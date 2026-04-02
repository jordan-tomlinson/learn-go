# learn-go
Learning Golang

## Race Conditions in Go

A **race condition** happens when multiple goroutines access shared data at the same time, and at least one of them is writing. The result becomes unpredictable because you can't control which goroutine runs first.

This repo contains two files that demonstrate the problem and the fix:

- `race_condition.go` — the broken version
- `race_condition_fix.go` — the fixed version

---

### The Problem (`race_condition.go`)

The broken code spawns multiple goroutines that all try to `append()` to the **same shared slice** at the same time:

```go
func processData(wg *sync.WaitGroup, result *[]int, data int) {
    defer wg.Done()
    *result = append(*result, data*2) // ❌ multiple goroutines do this simultaneously
}
```

#### Why `append()` is unsafe here

In Go, a slice is a small struct with three fields: a **pointer** to the underlying array, a **length**, and a **capacity**.

When you call `append()`, Go:
1. Reads the current length and capacity
2. Checks if there's room in the underlying array
3. If not, allocates a new, larger array and copies the data
4. Writes the value at the next position
5. Updates the length (and possibly the pointer and capacity)

None of these steps are atomic. When multiple goroutines run `append()` at the same time, they can:

- **Read the same length** — two goroutines both see `len = 2`, both write to index 2, and one value is lost
- **Trigger separate reallocations** — two goroutines both decide the array is full, both allocate new arrays, and one copy gets thrown away
- **Corrupt the slice header** — one goroutine updates the pointer while another is still writing to the old array

The result: missing values, duplicate values, out-of-order values, or even a crash.

```
// Possible outputs from the broken version:
[2 4 6 8 10]   // lucky — happened to work, but not guaranteed
[2 6 4 10 8]   // out of order
[2 4 4 8]      // missing values — data was silently lost
```

You can prove this to yourself by running Go's built-in race detector:

```sh
go run -race race_condition.go
```

This will print `WARNING: DATA RACE` and show you exactly which goroutines are conflicting.

---

### The Fix (`race_condition_fix.go`)

The fix makes **two key changes**:

#### 1. Preallocate the slice

```go
result := make([]int, len(input)) // creates [0, 0, 0, 0, 0]
```

Instead of starting with an empty slice and growing it with `append()`, we allocate the full slice upfront. The length and capacity are fixed — no goroutine will ever need to resize or reallocate.

#### 2. Write by index, not by append

```go
func processData(wg *sync.WaitGroup, result []int, index int, data int) {
    defer wg.Done()
    result[index] = data * 2 // each goroutine writes to its own unique index
}
```

Each goroutine receives its own `index` and writes only to that position. No two goroutines ever write to the same memory location, so there is no conflict — no lock or mutex needed.

#### Why this is safe

- The slice is **never resized** — no reallocation, no moving data
- Each goroutine writes to a **unique index** — no overlapping writes
- The slice header (pointer, length, capacity) is **never modified** — only the contents of the underlying array change, at distinct positions

```
// Output from the fixed version (always deterministic):
[2 4 6 8 10]
```

You can verify there are no races:

```sh
go run -race race_condition_fix.go
```

No warnings — clean.

---

### Quick Summary

| | Broken | Fixed |
|---|---|---|
| **Slice init** | `result := []int{}` (empty) | `result := make([]int, len(input))` (preallocated) |
| **Write method** | `append()` via pointer | Direct index assignment |
| **Shared mutation** | Yes — slice header + contents | No — each goroutine owns its index |
| **Thread-safe** | No | Yes |

---

### Other Ways to Fix Race Conditions

The index-based approach used here is the simplest fix for this pattern, but Go offers other tools:

- **`sync.Mutex`** — wrap the `append()` call in `Lock()` / `Unlock()` so only one goroutine appends at a time. Works, but adds contention.
- **Channels** — have each goroutine send its result to a channel, and collect values in the main goroutine. More idiomatic Go for many patterns.

The preallocated-slice approach is preferred here because it avoids locking entirely and is the most performant option when you know the size of the output ahead of time.
