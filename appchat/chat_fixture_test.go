package appchat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// docAPICall records one wedoc request as received by the fake server.
type docAPICall struct {
	Path  string
	Token string
	Body  map[string]any
}

// fakeDocAPI is an httptest replacement for the WeCom origin that serves
// gettoken and delegates every other path to a scripted handler.
type fakeDocAPI struct {
	server *httptest.Server

	mu         sync.Mutex
	tokenCalls int
	tokens     []string
	calls      []docAPICall

	handler func(index int, call docAPICall) map[string]any
}

func newFakeDocAPI(t *testing.T, handler func(index int, call docAPICall) map[string]any) *fakeDocAPI {
	t.Helper()
	fake := &fakeDocAPI{tokens: []string{"access-token-1"}, handler: handler}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeDocAPI) handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/cgi-bin/gettoken" {
		f.mu.Lock()
		f.tokenCalls++
		call := f.tokenCalls
		token := f.tokens[len(f.tokens)-1]
		if call-1 < len(f.tokens) {
			token = f.tokens[call-1]
		}
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0, "errmsg": "ok", "access_token": token, "expires_in": 7200,
		})
		return
	}

	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	call := docAPICall{Path: r.URL.Path, Token: r.URL.Query().Get("access_token"), Body: body}

	f.mu.Lock()
	f.calls = append(f.calls, call)
	index := len(f.calls)
	handler := f.handler
	f.mu.Unlock()

	response := map[string]any{"errcode": 0, "errmsg": "ok"}
	if handler != nil {
		response = handler(index, call)
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (f *fakeDocAPI) counts() (tokenCalls, endpointCalls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokenCalls, len(f.calls)
}

func (f *fakeDocAPI) recorded() []docAPICall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]docAPICall(nil), f.calls...)
}

func objectField(t *testing.T, value any, key string) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s is %T, want object", key, value)
	}
	return object
}

func arrayField(t *testing.T, value any, key string) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("%s is %T, want array", key, value)
	}
	return array
}
