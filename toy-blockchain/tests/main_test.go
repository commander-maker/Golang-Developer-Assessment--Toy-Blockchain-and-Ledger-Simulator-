package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to identify current test file")
	}
	return filepath.Dir(filepath.Dir(currentFile))
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "./cmd"}, args...)...)
	cmd.Dir = repoRoot(t)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestGenerateKeyHelperCreatesPemFile(t *testing.T) {
	name := "cli-test-key"
	keyPath := filepath.Join(repoRoot(t), name+".pem")
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove previous key file: %v", err)
	}
	defer os.Remove(keyPath)

	output, err := runCLI(t, "generate-key", name)
	if err != nil {
		t.Fatalf("expected generate-key to succeed, output: %s", output)
	}
	if !strings.Contains(output, "generated") {
		t.Fatalf("expected generation output, got: %s", output)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("expected PEM key file to exist: %v", err)
	}
}

func TestAddtxRejectsSenderKeyMismatch(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "cli-chain.json")
	keyPath := filepath.Join(repoRoot(t), "alice.pem")
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("expected alice.pem to exist for this CLI test: %v", err)
	}

	if output, err := runCLI(t, "init", "--file", dataFile); err != nil {
		t.Fatalf("expected init to succeed, output: %s", output)
	}
	if output, err := runCLI(t, "addtx", "SYSTEM", "Alice", "100", "--file", dataFile); err != nil {
		t.Fatalf("expected SYSTEM transaction to succeed, output: %s", output)
	}

	output, err := runCLI(t, "addtx", "Bob", "Alice", "60", "--key", keyPath, "--file", dataFile)
	if err == nil {
		t.Fatal("expected mismatched sender/key transaction to fail")
	}
	if !strings.Contains(output, "must match") {
		t.Fatalf("expected sender/key mismatch message, got: %s", output)
	}
}
