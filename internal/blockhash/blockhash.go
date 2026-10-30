// Package blockhash computes the cumulative block SHA that the WeCom micro disk
// and document upload endpoints expect for chunked uploads. The appendix specifies
// a mix that the standard library cannot express: every block but the last reports
// the SHA-1 chaining value after hashing the bytes so far, while the last reports
// the digest of the whole file.
package blockhash

import (
	"encoding/hex"
	"math/bits"
)

// SHA1 returns one cumulative SHA per block of size bytes, in order. An empty
// input yields no blocks.
func SHA1(content []byte, size int) []string {
	if len(content) == 0 || size <= 0 {
		return nil
	}
	var state sha1State
	state.init()
	sums := make([]string, 0, (len(content)+size-1)/size)
	for offset := 0; offset < len(content); offset += size {
		end := offset + size
		if end > len(content) {
			end = len(content)
		}
		state.update(content[offset:end])
		if end == len(content) {
			sums = append(sums, hex.EncodeToString(state.sum()))
			continue
		}
		sums = append(sums, hex.EncodeToString(state.digest()))
	}
	return sums
}

// sha1State is the 160-bit SHA-1 chaining value plus the bytes of the current
// partial block. The standard library finalises a digest and never exposes the
// intermediate value, but the upload appendix asks for it, so the compression
// function lives here.
type sha1State struct {
	h       [5]uint32
	buf     [64]byte
	pending int
	total   uint64
}

func (s *sha1State) init() {
	s.h = [5]uint32{0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0}
}

func (s *sha1State) update(content []byte) {
	s.total += uint64(len(content))
	if s.pending > 0 {
		copied := copy(s.buf[s.pending:], content)
		s.pending += copied
		content = content[copied:]
		if s.pending < len(s.buf) {
			return
		}
		s.compress(s.buf[:])
		s.pending = 0
	}
	for len(content) >= len(s.buf) {
		s.compress(content[:len(s.buf)])
		content = content[len(s.buf):]
	}
	if len(content) > 0 {
		s.pending = copy(s.buf[:], content)
	}
}

func (s *sha1State) compress(block []byte) {
	var words [80]uint32
	for i := 0; i < 16; i++ {
		words[i] = uint32(block[i*4])<<24 | uint32(block[i*4+1])<<16 |
			uint32(block[i*4+2])<<8 | uint32(block[i*4+3])
	}
	for i := 16; i < 80; i++ {
		words[i] = bits.RotateLeft32(words[i-3]^words[i-8]^words[i-14]^words[i-16], 1)
	}
	a, b, c, d, e := s.h[0], s.h[1], s.h[2], s.h[3], s.h[4]
	for i := 0; i < 80; i++ {
		var f, k uint32
		switch {
		case i < 20:
			f, k = (b&c)|(^b&d), 0x5A827999
		case i < 40:
			f, k = b^c^d, 0x6ED9EBA1
		case i < 60:
			f, k = (b&c)|(b&d)|(c&d), 0x8F1BBCDC
		default:
			f, k = b^c^d, 0xCA62C1D6
		}
		a, b, c, d, e = bits.RotateLeft32(a, 5)+f+e+k+words[i], a,
			bits.RotateLeft32(b, 30), c, d
	}
	s.h[0] += a
	s.h[1] += b
	s.h[2] += c
	s.h[3] += d
	s.h[4] += e
}

// digest renders the current chaining value, which is the intermediate value the
// appendix calls the accumulated SHA.
func (s *sha1State) digest() []byte {
	out := make([]byte, 20)
	for i, word := range s.h {
		out[i*4] = byte(word >> 24)
		out[i*4+1] = byte(word >> 16)
		out[i*4+2] = byte(word >> 8)
		out[i*4+3] = byte(word)
	}
	return out
}

// sum finalises a copy of the state: append the 0x80 marker, pad to a 64-byte
// boundary and append the message length in bits.
func (s sha1State) sum() []byte {
	total := s.total
	s.update([]byte{0x80})
	for s.pending != 56 {
		s.update([]byte{0})
	}
	var length [8]byte
	messageBits := total << 3
	for i := 7; i >= 0; i-- {
		length[i] = byte(messageBits)
		messageBits >>= 8
	}
	s.update(length[:])
	return s.digest()
}
