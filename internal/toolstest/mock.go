package toolstest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/itokun99/blogger-go"
)

type RecordedCall struct {
	Method   string
	Path     string
	RawQuery string
	Body     []byte
}

type MockAPI struct {
	t      *testing.T
	server *httptest.Server
	stubs  map[string]stubResponse
	calls  []RecordedCall
}

type stubResponse struct {
	status int
	body   string
}

func NewMockAPI(t *testing.T) *MockAPI {
	m := &MockAPI{
		t:     t,
		stubs: make(map[string]stubResponse),
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handler))
	t.Cleanup(m.server.Close)
	return m
}

func (m *MockAPI) Stub(method, path string, status int, body string) {
	key := method + " " + path
	m.stubs[key] = stubResponse{status: status, body: body}
}

func (m *MockAPI) Calls() []RecordedCall {
	out := make([]RecordedCall, len(m.calls))
	for i := range m.calls {
		out[i] = m.calls[i]
	}
	return out
}

func (m *MockAPI) Client() *blogger.Client {
	httpClient := &http.Client{
		Transport: &mockTransport{mock: m},
	}
	return blogger.NewRawClient(httpClient)
}

type mockTransport struct {
	mock *MockAPI
}

func (mt *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	mt.mock.calls = append(mt.mock.calls, RecordedCall{
		Method:   req.Method,
		Path:     req.URL.Path,
		RawQuery: req.URL.RawQuery,
		Body:     bodyBytes,
	})

	return mt.mock.handlerResponse(req)
}

func (m *MockAPI) handler(w http.ResponseWriter, r *http.Request) {
	resp, err := m.handlerResponse(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	}
}

func (m *MockAPI) handlerResponse(r *http.Request) (*http.Response, error) {
	key := r.Method + " " + r.URL.Path
	stub, ok := m.stubs[key]
	if !ok {
		errBody := `{"error":{"code":404,"message":"Not stubbed: ` + r.Method + " " + r.URL.Path + `"}}`
		return &http.Response{
			StatusCode: 404,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(errBody)),
		}, nil
	}
	return &http.Response{
		StatusCode: stub.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(stub.body)),
	}, nil
}

func Ptr[T any](v T) *T {
	return &v
}
