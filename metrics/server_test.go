package metrics

import (
	"bytes"
	metricsmocks "github.com/chainalysis-oss/oslc/mocks/oslc/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_Serve(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	s := &Server{
		options: &serverOptions{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		httpServer: &http.Server{},
	}

	go func() {
		err = s.Serve(listener)
		require.NoError(t, err)
	}()

	client := &http.Client{}
	resp, err := client.Get("http://" + listener.Addr().String() + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestServer_Close(t *testing.T) {
	s := &Server{
		options: &serverOptions{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		httpServer: &http.Server{},
	}

	err := s.Close()
	require.NoError(t, err)
}

func TestServer_Close_Error(t *testing.T) {
	mock := metricsmocks.NewMockhttpServer(t)
	mock.EXPECT().Close().Return(io.EOF)
	s := &Server{
		options: &serverOptions{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		httpServer: mock,
	}

	err := s.Close()
	require.ErrorIs(t, err, io.EOF)
}

func TestServer_GetPrometheusRegistry(t *testing.T) {
	s := &Server{
		options: &serverOptions{
			PrometheusRegistry: prometheus.NewRegistry(),
		},
	}

	require.Equal(t, s.options.PrometheusRegistry, s.GetPrometheusRegistry())
}

func TestHttpLoggerMiddleware(t *testing.T) {
	logs := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(logs, nil))
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := httpLoggerMiddleware(logger, nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	require.NotEmpty(t, logs.String())
}
