package http_retry

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

func Test_getBody_CanRepeatedlyRead(t *testing.T) {
	rs, err := os.Open("test.txt")
	require.NoError(t, err)
	r, err := http.NewRequest(http.MethodPut, "http://localhost", rs)
	require.NoError(t, err)
	getBodyFunc := getBody(r)

	result := make([]byte, 3)
	for i := 0; i < 2; i++ {
		body, err := getBodyFunc()
		require.NoError(t, err)
		n, err := body.Read(result)
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, "abc", string(result))
	}
}

func Test_getBody_CanRepeatedlyReadFromMiddleway(t *testing.T) {
	rs, err := os.Open("test.txt")
	require.NoError(t, err)
	rs.Seek(1, io.SeekStart)
	r, err := http.NewRequest(http.MethodPut, "http://localhost", rs)
	require.NoError(t, err)
	getBodyFunc := getBody(r)

	result := make([]byte, 2)
	for i := 0; i < 2; i++ {
		body, err := getBodyFunc()
		require.NoError(t, err)
		n, err := body.Read(result)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.Equal(t, "bc", string(result))
	}
	position, err := rs.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	assert.Equal(t, int64(3), position)
}