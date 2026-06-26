//go:build (!cgo || nocgo) && !unix

package fitz

// captureStderr runs fn directly where redirecting the C runtime's stderr isn't supported.
func captureStderr(fn func()) string {
	fn()
	return ""
}
