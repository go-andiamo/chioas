package codegen

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriter_writeIndent(t *testing.T) {
	t.Run("max tabs ok", func(t *testing.T) {
		require.True(t, maxT > 1)
	})
	t.Run("none", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(0))
		require.NoError(t, w.err)
		require.Equal(t, "", buf.String())
	})
	t.Run("one", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(1))
		require.NoError(t, w.err)
		require.Equal(t, "\t", buf.String())
	})
	t.Run("two", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(2))
		require.NoError(t, w.err)
		require.Equal(t, "\t\t", buf.String())
	})
	t.Run("eleven", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(11))
		require.NoError(t, w.err)
		require.Equal(t, "\t\t\t\t\t\t\t\t\t\t\t", buf.String())
	})
	t.Run("boundary", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(maxT))
		require.NoError(t, w.err)
		require.Len(t, buf.Bytes(), maxT)
	})
	t.Run("forty two", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		require.True(t, w.writeIndent(42))
		require.NoError(t, w.err)
		require.Len(t, buf.Bytes(), 42)
	})
}

func TestCodeWriter_writeLf(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		w.writeLf(false)
		require.NoError(t, w.err)
		require.Equal(t, "\n", buf.String())
	})
	t.Run("double", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, false)
		w.writeLf(true)
		require.NoError(t, w.err)
		require.Equal(t, "\n\n", buf.String())
	})
	t.Run("crlf", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, true)
		w.writeLf(false)
		require.NoError(t, w.err)
		require.Equal(t, "\r\n", buf.String())
	})
	t.Run("double crlf", func(t *testing.T) {
		var buf bytes.Buffer
		w := newWriter(&buf, false, true)
		w.writeLf(true)
		require.NoError(t, w.err)
		require.Equal(t, "\r\n\r\n", buf.String())
	})
}

func TestCodeWriter_formattingFailed(t *testing.T) {
	var buf bytes.Buffer
	w := newWriter(&buf, true, false)
	_, _ = w.w.Write([]byte("this is bad syntax"))
	err := w.format()
	require.Error(t, err)
}
