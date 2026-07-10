package tests

import (
	"os"
	"path/filepath"
	"testing"
	"toy-blockchain/blockchain"
)

func TestLoadFromFile_EmptyFileCreatesNewBlockchain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	bc, err := blockchain.LoadFromFile(path)
	if err != nil {
		t.Fatalf("expected no error loading empty file, got %v", err)
	}

	if len(bc.Blocks) != 1 {
		t.Fatalf("expected genesis block after loading empty file, got %d blocks", len(bc.Blocks))
	}
	if bc.Blocks[0].Index != 0 {
		t.Fatalf("expected genesis block index 0, got %d", bc.Blocks[0].Index)
	}
}
