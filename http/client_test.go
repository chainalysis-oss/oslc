package http

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func TestClient_Query(t *testing.T) {
	testcases := []struct {
		name              string
		url               string
		want              int
		wantErr           bool
		mockFunc          RoundTripFunc
		additionalOptions []ClientOption
	}{
		{
			name:    "success - no resp body",
			url:     "https://example.com",
			want:    http.StatusOK,
			wantErr: false,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			},
		},
		{
			name:    "success - resp body",
			url:     "https://example.com",
			want:    http.StatusOK,
			wantErr: false,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`))),
				}, nil
			},
		},
		{
			name:    "failure - invalid url",
			url:     "://example.com",
			wantErr: true,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			},
		},
		{
			name: "debug logging enabled",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			},
			additionalOptions: []ClientOption{
				WithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(make([]byte, 0)), &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}))),
			},
			want:    http.StatusOK,
			wantErr: false,
			url:     "https://example.com",
		},
		{
			name: "failure - error during request execution",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return nil, assert.AnError
			},
			wantErr: true,
			url:     "https://example.com",
		},
		{
			name: "failure - error during response body read",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(errReader(0)),
				}, nil
			},
			wantErr: true,
			url:     "https://example.com",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewTestHTTPClient(tt.mockFunc)
			tt.additionalOptions = append(tt.additionalOptions, WithHTTPClient(mock))
			c, err := NewClient(tt.additionalOptions...)
			require.NoError(t, err)
			resp, err := c.Query(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, resp.StatusCode)
		})
	}
}

func BenchmarkClient_Query_debug_logging(b *testing.B) {
	mock := NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`))),
		}, nil
	})
	c, err := NewClient(WithHTTPClient(mock), WithLogger(slog.New(slog.NewTextHandler(bytes.NewBuffer(make([]byte, 0)), &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Query("https://example.com")
		require.NoError(b, err)
	}
}

func BenchmarkClient_Query_no_debug_logging(b *testing.B) {
	mock := NewTestHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"key":"value"}`))),
		}, nil
	})
	c, err := NewClient(WithHTTPClient(mock))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Query("https://example.com")
		require.NoError(b, err)
	}
}
