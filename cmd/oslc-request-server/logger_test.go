package main

import (
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func Test_logLevelFromStr(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "debug",
			args: args{
				level: "debug",
			},
			want: "debug",
		},
		{
			name: "info",
			args: args{
				level: "info",
			},
			want: "info",
		},
		{name: "warn",
			args: args{
				level: "warn",
			},
			want: "warn",
		},
		{
			name: "error",
			args: args{
				level: "error",
			},
			want: "error",
		},
		{
			name: "invalid",
			args: args{
				level: "invalid",
			},
			want: "info",
		},
	}
	for _, tt := range tests {
		lvl := logLevelFromStr(tt.args.level)
		require.Equal(t, strings.ToLower(lvl.String()), tt.want)
	}
}

func Test_getLogger(t *testing.T) {
	type args struct {
		level string
		kind  string
	}
	tests := []struct {
		name string
		args args
		want *slog.Logger
	}{
		{
			name: "text",
			args: args{
				level: "info",
				kind:  "text",
			},
			want: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})),
		},
		{
			name: "json",
			args: args{
				level: "warn",
				kind:  "json",
			},
			want: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelWarn,
			})),
		},
		{
			name: "invalid",
			args: args{
				level: "info",
				kind:  "invalid",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		logger := getLogger(tt.args.level, tt.args.kind, io.Discard)
		require.Equal(t, logger, tt.want)
	}
}
