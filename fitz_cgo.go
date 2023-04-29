//go:build !extlib

package fitz

/*
#cgo CFLAGS: -Iinclude

#cgo linux,386 LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_386 -lmupdfthird_linux_386 -lm
#cgo linux,amd64,!musl LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_amd64 -lmupdfthird_linux_amd64 -lm
#cgo linux,amd64,musl LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_amd64_musl -lmupdfthird_linux_amd64_musl -lm
#cgo linux,!android,arm LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_arm -lmupdfthird_linux_arm -lm
#cgo linux,!android,arm64,!musl LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_arm64 -lmupdfthird_linux_arm64 -lm
#cgo linux,!android,arm64,musl LDFLAGS: -L${SRCDIR}/libs -lmupdf_linux_arm64_musl -lmupdfthird_linux_arm64_musl -lm
#cgo android,arm LDFLAGS: -L${SRCDIR}/libs -lmupdf_android_arm -lmupdfthird_android_arm -lm -llog
#cgo android,arm64 LDFLAGS: -L${SRCDIR}/libs -lmupdf_android_arm64 -lmupdfthird_android_arm64 -lm -llog
#cgo windows,386 LDFLAGS: -L${SRCDIR}/libs -lmupdf_windows_386 -lmupdfthird_windows_386 -lm -lcomdlg32 -lgdi32 -lmsvcr90 -Wl,--allow-multiple-definition
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/libs -lmupdf_windows_amd64 -lmupdfthird_windows_amd64 -lm -lcomdlg32 -lgdi32 -Wl,--allow-multiple-definition
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/libs -lmupdf_darwin_amd64 -lmupdfthird_darwin_amd64 -lm
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/libs -lmupdf_darwin_arm64 -lmupdfthird_darwin_arm64 -lm
*/
import "C"
