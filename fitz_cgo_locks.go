//go:build cgo && !nocgo

package fitz

/*
#include <mupdf/fitz.h>

#if defined(_WIN32)
#include <windows.h>
static SRWLOCK mutexes[FZ_LOCK_MAX];
#else
#include <pthread.h>
static pthread_mutex_t mutexes[FZ_LOCK_MAX];
#endif

static void lock_mutex(void *user, int lock) {
#if defined(_WIN32)
	AcquireSRWLockExclusive(&mutexes[lock]);
#else
	pthread_mutex_lock(&mutexes[lock]);
#endif
}

static void unlock_mutex(void *user, int lock) {
#if defined(_WIN32)
	ReleaseSRWLockExclusive(&mutexes[lock]);
#else
	pthread_mutex_unlock(&mutexes[lock]);
#endif
}

static fz_locks_context locks = { NULL, lock_mutex, unlock_mutex };

static fz_locks_context *init_locks(void) {
	for (int i = 0; i < FZ_LOCK_MAX; i++) {
#if defined(_WIN32)
		InitializeSRWLock(&mutexes[i]);
#else
		if (pthread_mutex_init(&mutexes[i], NULL) != 0) {
			while (i > 0)
				pthread_mutex_destroy(&mutexes[--i]);
			return NULL;
		}
#endif
	}
	return &locks;
}

static fz_context *new_context(fz_locks_context *locks, size_t max_store) {
	return fz_new_context(NULL, locks, max_store);
}
*/
import "C"

// Initialize once during package initialization. The native mutexes live for
// the process lifetime and are shared by all contexts, including independent
// documents. MuPDF calls these handlers without crossing back into Go.
var contextLocks = C.init_locks()

func newContext() *C.fz_context {
	if contextLocks == nil {
		return nil
	}
	return C.new_context(contextLocks, C.size_t(MaxStore))
}
