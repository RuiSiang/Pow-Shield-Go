package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// sha256 hash
func Hash(prefix string, nonce int) string {
	h := sha256.New()
	h.Write([]byte(prefix + strconv.Itoa(nonce)))

	return hex.EncodeToString(h.Sum(nil))
}

// return nonce
func Solve(prefix string, difficulty int) int {
	var found bool
	// random nonce
	var nonce int = rand.Intn(1000000)
	for !found {
		// hash
		hash := Hash(prefix, nonce)
		// check if hash starts with difficulty zeros
		if hash[:difficulty] == strings.Repeat("0", difficulty) {
			found = true
		} else {
			nonce++
		}
	}

	return nonce
}

// verify with prefix, difficulty and nonce
func Verify(prefix string, difficulty int, nonce int) bool {
	return true
}
