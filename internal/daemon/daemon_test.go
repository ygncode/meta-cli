package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPIDPath(t *testing.T) {
	got := PIDPath("/tmp/test")
	want := filepath.Join("/tmp/test", "webhook.pid")
	if got != want {
		t.Errorf("PIDPath = %q, want %q", got, want)
	}
}

func TestLogPath(t *testing.T) {
	got := LogPath("/tmp/test")
	want := filepath.Join("/tmp/test", "webhook.log")
	if got != want {
		t.Errorf("LogPath = %q, want %q", got, want)
	}
}

func TestWriteReadPID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pid")
	if err := WritePID(path, 12345); err != nil {
		t.Fatalf("WritePID: %v", err)
	}
	pid, err := ReadPID(path)
	if err != nil {
		t.Fatalf("ReadPID: %v", err)
	}
	if pid != 12345 {
		t.Errorf("ReadPID = %d, want 12345", pid)
	}
}

func TestReadPIDMissing(t *testing.T) {
	_, err := ReadPID(filepath.Join(t.TempDir(), "nope.pid"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIsRunningCurrentProcess(t *testing.T) {
	if !IsRunning(os.Getpid()) {
		t.Error("expected current process to be running")
	}
}

func TestIsRunningDeadPID(t *testing.T) {
	// PID 2^20 is very unlikely to exist
	if IsRunning(1 << 20) {
		t.Skip("PID 1048576 unexpectedly exists")
	}
}

func TestRemovePID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.pid")
	if err := os.WriteFile(path, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RemovePID(path); err != nil {
		t.Fatalf("RemovePID: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestRemovePIDMissing(t *testing.T) {
	err := RemovePID(filepath.Join(t.TempDir(), "nope.pid"))
	if err != nil {
		t.Fatalf("RemovePID on missing file: %v", err)
	}
}
