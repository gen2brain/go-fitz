//go:build cgo && !nocgo

package fitz_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	"github.com/gen2brain/go-fitz"
)

const (
	stressWorkers    = 8
	stressIterations = 20
)

func runConcurrentImageDPIStress(t *testing.T) {
	t.Helper()

	pdfBytes, err := os.ReadFile(filepath.Join("testdata", "concurrency.pdf"))
	if err != nil {
		t.Fatalf("you need to provide a concurrency.pdf: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, stressWorkers)

	for worker := 0; worker < stressWorkers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for iter := 0; iter < stressIterations; iter++ {
				doc, err := fitz.NewFromMemory(pdfBytes)
				if err != nil {
					errCh <- fmt.Errorf("worker %d iteration %d open doc: %w", worker, iter, err)
					return
				}

				page := iter % doc.NumPage()
				img, err := doc.ImageDPI(page, 144)
				closeErr := doc.Close()
				if err != nil {
					errCh <- fmt.Errorf("worker %d iteration %d render: %w", worker, iter, err)
					return
				}
				if closeErr != nil {
					errCh <- fmt.Errorf("worker %d iteration %d close: %w", worker, iter, closeErr)
					return
				}
				if img == nil {
					errCh <- fmt.Errorf("worker %d iteration %d render: nil image", worker, iter)
					return
				}
			}
		}(worker)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestConcurrentImageDPIStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skip stress test in short mode")
	}

	runConcurrentImageDPIStress(t)
}

func TestConcurrentImageDPISubprocessNoSegv(t *testing.T) {
	if os.Getenv("GOFITZ_CONCURRENCY_CHILD") == "1" {
		runConcurrentImageDPIStress(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestConcurrentImageDPISubprocessNoSegv$")
	cmd.Env = append(os.Environ(), "GOFITZ_CONCURRENCY_CHILD=1")

	output, err := cmd.CombinedOutput()
	if err == nil {
		return
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			t.Fatalf("child process terminated by signal %s\n%s", status.Signal(), output)
		}
	}

	t.Fatalf("child process failed: %v\n%s", err, output)
}
