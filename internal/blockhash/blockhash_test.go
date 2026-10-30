package blockhash

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

// blockSize mirrors the protocol block size used by the upload endpoints.
const blockSize = 2 << 20

// sha1Hex is the reference digest the last block must match.
func sha1Hex(content []byte) string {
	sum := sha1.Sum(content)
	return hex.EncodeToString(sum[:])
}

func TestBlockSHALastValueIsTheFileDigest(t *testing.T) {
	cases := map[string]int{
		"one byte":           1,
		"short":              100,
		"block minus one":    blockSize - 1,
		"exactly one block":  blockSize,
		"one block plus one": blockSize + 1,
		"three blocks":       3 * blockSize,
		"three plus a tail":  3*blockSize + 12345,
	}
	for name, size := range cases {
		t.Run(name, func(t *testing.T) {
			content := make([]byte, size)
			for i := range content {
				content[i] = byte(i * 7)
			}
			sums := SHA1(content, blockSize)
			wantBlocks := (size + blockSize - 1) / blockSize
			if len(sums) != wantBlocks {
				t.Fatalf("blocks = %d, want %d", len(sums), wantBlocks)
			}
			if got, want := sums[len(sums)-1], sha1Hex(content); got != want {
				t.Fatalf("last block = %s, want %s", got, want)
			}
			for i, sum := range sums {
				if len(sum) != 40 {
					t.Fatalf("block %d has %d hex chars, want 40", i, len(sum))
				}
				if _, err := hex.DecodeString(sum); err != nil {
					t.Fatalf("block %d is not hex: %v", i, err)
				}
			}
		})
	}
	if sums := SHA1(nil, blockSize); len(sums) != 0 {
		t.Fatalf("empty content returned %d blocks", len(sums))
	}
}

// TestBlockSHAIntermediateStatesResume proves the intermediate values really are
// SHA-1 chaining values: resuming from one with the remaining bytes must produce
// the digest of the whole file. A prefix digest would fail this.
func TestBlockSHAIntermediateStatesResume(t *testing.T) {
	content := make([]byte, 2*blockSize+9999)
	for i := range content {
		content[i] = byte(i * 31)
	}
	sums := SHA1(content, blockSize)
	if len(sums) != 3 {
		t.Fatalf("blocks = %d, want 3", len(sums))
	}
	want := sha1Hex(content)
	for index := 0; index < len(sums)-1; index++ {
		raw, err := hex.DecodeString(sums[index])
		if err != nil {
			t.Fatalf("block %d: %v", index, err)
		}
		var state sha1State
		for word := range state.h {
			state.h[word] = uint32(raw[word*4])<<24 | uint32(raw[word*4+1])<<16 |
				uint32(raw[word*4+2])<<8 | uint32(raw[word*4+3])
		}
		consumed := (index + 1) * blockSize
		state.total = uint64(consumed)
		state.update(content[consumed:])
		if got := hex.EncodeToString(state.sum()); got != want {
			t.Fatalf("state after block %d is not a chaining value: %s, want %s", index, got, want)
		}
	}
}
