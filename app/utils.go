package app

import (
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// vec2 returns a new [rl.Vector2] (shorthand for [rl.NewVector2])
func vec2(x, y float32) rl.Vector2 { return rl.Vector2{X: x, Y: y} }

// assert panics if the condition is false
func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

// Range returns a slice of integers [i; j[
func Range(i, j int) []int {
	r := make([]int, j-i)
	for k := range r {
		r[k] = i + k
	}
	return r
}

// Returns a new slice with the elements at idxs
func CopyIdxs[T any](s []T, idxs []int) []T {
	dst := make([]T, len(idxs))
	for i, idx := range idxs {
		dst[i] = s[idx]
	}
	return dst
}

// SwapDelete efficiently deletes the element at index i by swapping it with the last element
// and then truncating the slice.
//
// Calling [SwapDelete] with i == len(s) - 1 simply removes the last element.
//
// The complexity is O(1) but the slice order is not preserved.
func SwapDelete[T any](s []T, i int) []T {
	last := len(s) - 1
	if i < last {
		s[i], s[last] = s[last], s[i]
	}
	return s[:last]
}

// SwapInsert efficiently inserts the given element at index i by pushing the current i-th element
// to the back.
//
// Calling [SwapInsert] with i == len(s) simply appends the element.
//
// This is the reverse of [SwapDelete].
//
// The complexity is O(1) but the slice order is not preserved.
func SwapInsert[T any](s []T, i int, v T) []T {
	n := len(s)
	if i == n {
		s = append(s, v)
		return s
	}
	s = append(s, s[i])
	s[i] = v
	return s
}

// SwapDeleteMany efficiently deletes the elements at given indices idxs by swapping them with the last elements
// and then truncating the slice.
//
// idxs must be sorted in ascending order.
//
// The complexity is O(len(indices)) but the slice order is not preserved.
func SwapDeleteMany[T any](s []T, idxs []int) []T {
	// reverse order because we are deleting elements
	for i := len(idxs) - 1; i >= 0; i-- {
		s = SwapDelete(s, idxs[i])
	}
	return s
}

// SwapInsertMany efficiently inserts the given elements at given indices idxs by pushing the current$
// elements at idxs to the back.
//
// Idxs values must all be between 0 and len(s)+len(idxs)-1.
//
// This is the reverse of [SwapDeleteMany].
//
// The complexity is O(len(indices)) but the slice order is not preserved.
func SwapInsertMany[T any](s []T, idxs []int, vs []T) []T {
	for i := 0; i < len(idxs); i++ {
		s = SwapInsert(s, idxs[i], vs[i])
	}
	return s
}

// SortedIntsIndex returns the index of x in a ascending sorted int slice (-1 if not found)
//
// Binary search is used.
func SortedIntsIndex(a []int, x int) int {
	i, j := 0, len(a)
	if len(a) == 0 {
		return -1
	}
	for i < j {
		k := int(uint(i+j) >> 1) // avoid overflow when computing h
		if a[k] == x {
			return k
		} else if a[k] < x {
			i = k + 1
		} else {
			j = k
		}
	}
	if i < len(a) && a[i] == x {
		println("!!! sortedIntsIndex: found after loop\n")
		return i
	}
	return -1
}

// NormalizePath normalizes a path to use the OS path separator.
//
// It preserves the trailing separator if any.
func NormalizePath(p string) string {
	ps := strings.Split(p, string(os.PathListSeparator))
	for i, p := range ps {
		p = filepath.ToSlash(p)
		endSep := len(p) > 0 && p[len(p)-1] == '/'
		p = filepath.FromSlash(p)
		p = filepath.Clean(p)
		if endSep {
			ps[i] = p + string(os.PathSeparator)
		} else {
			ps[i] = p
		}
	}
	return strings.Join(ps, string(os.PathListSeparator))
}

// Replaces single and double quotes with asterisks.
func RemoveQuotes(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `"`, "*"), `'`, "*")
}

func CheckCollisionRecLine(rec rl.Rectangle, p1, p2 rl.Vector2) bool {
	tl, tr, bl, br := rec.TopLeft(), rec.TopRight(), rec.BottomLeft(), rec.BottomRight()
	var p rl.Vector2
	return rec.CheckCollisionPoint(p1) || rec.CheckCollisionPoint(p2) ||
		rl.CheckCollisionLines(p1, p2, tl, tr, &p) ||
		rl.CheckCollisionLines(p1, p2, bl, br, &p) ||
		rl.CheckCollisionLines(p1, p2, tl, bl, &p) ||
		rl.CheckCollisionLines(p1, p2, tr, br, &p)
}
