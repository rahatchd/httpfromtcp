package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidSingleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok := headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", hdr)
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestValidSingleHeaderWithExtraWhitespace(t *testing.T) {
	headers := NewHeaders()
	data := []byte(" Host:  localhost:42069   \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok := headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", hdr)
	assert.Equal(t, 28, n)
	assert.False(t, done)
}

func TestValidTwoHeadersWithExistingHeaders(t *testing.T) {
	headers := Headers{"host": "localhost:42069"}
	data := []byte("Agent-Type: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok := headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", hdr)
	hdr, ok = headers.Get("Agent-Type")
	assert.True(t, ok)
	assert.Equal(t, "curl/7.81.0", hdr)
	assert.Equal(t, 25, n)
	assert.False(t, done)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n extra stuff")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}

func TestInvalidSpacingHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestInvalidChar(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestMultipleValues(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Set-Person: lane-loves-go\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok := headers.Get("Set-Person")
	assert.True(t, ok)
	assert.Equal(t, "lane-loves-go", hdr)
	assert.Equal(t, 27, n)
	assert.False(t, done)
	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok = headers.Get("Set-Person")
	assert.True(t, ok)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", hdr)
	assert.Equal(t, 29, n)
	assert.False(t, done)
	data = []byte("Set-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	hdr, ok = headers.Get("Set-Person")
	assert.True(t, ok)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", hdr)
	assert.Equal(t, 28, n)
	assert.False(t, done)
}
