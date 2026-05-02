package snhash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashDirDeterministic(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b.txt"), []byte("bravo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "a.txt"), []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := HashDir(root, WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashDir(root, WithProfile(ProfileEvidence), WithCreatedAt(fixedTime))
	if err != nil {
		t.Fatal(err)
	}
	firstRoot, err := ProofRoot(first)
	if err != nil {
		t.Fatal(err)
	}
	secondRoot, err := ProofRoot(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstRoot.Hex() != secondRoot.Hex() {
		t.Fatalf("directory proof root not deterministic: %s != %s", firstRoot.Hex(), secondRoot.Hex())
	}
	if len(first.Entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(first.Entries))
	}
	if first.Entries[0].Path != "b.txt" || first.Entries[1].Path != "sub" || first.Entries[2].Path != "sub/a.txt" {
		t.Fatalf("entries not sorted by path: %#v", first.Entries)
	}
}
