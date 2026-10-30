package kf

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateUTF8KeepsValidText(t *testing.T) {
	long := strings.Repeat("测", 1000) // 3000 bytes
	got := truncateUTF8(long, maxTextBytes)
	if len(got) > maxTextBytes {
		t.Fatalf("truncated length = %d", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatal("truncation split a rune")
	}
	if short := "你好"; truncateUTF8(short, maxTextBytes) != short {
		t.Fatal("short text must be unchanged")
	}
}
