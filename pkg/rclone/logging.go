/*
Copyright 2025 Veloxpack.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rclone

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rclone/rclone/fs"
	"k8s.io/klog/v2"
)

// klogHandler implements slog.Handler to bridge rclone logs to klog
type klogHandler struct {
	attrs  []slog.Attr
	groups []string
}

// Enabled reports whether the handler handles records at the given level.
func (h *klogHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Always return true - we'll handle level filtering in Handle()
	return true
}

// Handle formats the log record and sends it to klog with appropriate verbosity
func (h *klogHandler) Handle(_ context.Context, record slog.Record) error {
	// Build the message with attributes
	var sb strings.Builder
	sb.WriteString(record.Message)

	// Add attributes if any
	if len(h.attrs) > 0 || record.NumAttrs() > 0 {
		sb.WriteString(" {")

		// Add handler-level attributes
		for i, attr := range h.attrs {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(attr.Key)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprint(attr.Value.Any()))
		}

		// Add record-level attributes
		first := len(h.attrs) == 0
		record.Attrs(func(attr slog.Attr) bool {
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(attr.Key)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprint(attr.Value.Any()))
			return true
		})

		sb.WriteString("}")
	}

	message := sb.String()

	// Map slog levels to klog levels
	// slog.Level mapping:
	// LevelDebug = -4
	// LevelInfo = 0
	// LevelWarn = 4
	// LevelError = 8
	switch {
	case record.Level >= slog.LevelError:
		// Error level -> klog.Error
		klog.ErrorDepth(2, message)
	case record.Level >= slog.LevelWarn:
		// Warning level -> klog.Warning
		klog.WarningDepth(2, message)
	case record.Level >= slog.LevelInfo:
		// Info level -> klog.Info (V(2))
		klog.V(2).InfoDepth(2, message)
	default:
		// Debug level -> klog.V(4)
		klog.V(4).InfoDepth(2, message)
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes
func (h *klogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &klogHandler{
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new handler with an additional group
func (h *klogHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &klogHandler{
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// InitRcloneLogging sets up rclone to log through klog
// This should be called during driver initialization before any rclone operations
func InitRcloneLogging() {
	// Create our custom klog handler
	handler := &klogHandler{}

	// Set it as the default slog handler that rclone will use
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Also set the rclone log level to match our verbosity
	// Start with INFO level - can be adjusted via klog flags
	fs.GetConfig(context.TODO()).LogLevel = fs.LogLevelInfo

	klog.V(2).Info("Rclone logging initialized - logs will be redirected to klog")
}
