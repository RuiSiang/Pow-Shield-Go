package pow

import "testing"

func TestHash(t *testing.T) {
	// test hash
	hash := Hash("prefix", 1)
	if hash != "47b90cc2773688f283878f067ad8c92eb774d8ea12b6f8b782560ed244dc7fd1" {
		t.Errorf("Hash(prefix, 1) = %s, want %s", hash, "47b90cc2773688f283878f067ad8c92eb774d8ea12b6f8b782560ed244dc7fd1")
	}
}

func TestSolve(t *testing.T) {
	// test solve
	nonce := Solve("prefix", 3)
	hash := Hash("prefix", nonce)
	if hash[:3] != "000" {
		t.Errorf("Hash(prefix, %d) = %s, want %s", nonce, hash, "000")
	}
}
