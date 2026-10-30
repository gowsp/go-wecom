package payment

import (
	"strings"
	"testing"
)

// TestEncodeEscapesCDATA keeps user-supplied text from terminating the CDATA
// section early, which would corrupt the XML body.
func TestEncodeEscapesCDATA(t *testing.T) {
	body := string(encode(map[string]string{"wishing": "谢谢]]>支持"}))
	if strings.Contains(body, "]]>支持") {
		t.Fatalf("CDATA terminator not escaped: %s", body)
	}
	if !strings.Contains(body, "]]]]><![CDATA[>") {
		t.Fatalf("unexpected escaping: %s", body)
	}
}
