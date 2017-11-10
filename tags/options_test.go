package http_ctxtags_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/improbable-eng/go-httpwares/tags"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServiceNameDetector(t *testing.T) {
	for _, tcase := range []struct {
		input    string
		expected string
	}{
		{
			input:    "www.googleapis.com",
			expected: "googleapis",
		},
		{
			input:    "autoscaling.us-east-2.amazonaws.com",
			expected: "amazonaws",
		},
		{
			input:    "www.googleapis.com:443",
			expected: "googleapis",
		},
		{
			input:    "whatever.net:1234",
			expected: "whatever",
		},
		{
			input:    "1.2.3.4:1234",
			expected: http_ctxtags.DefaultServiceName,
		},
	} {
		t.Run("for_"+tcase.input, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{
					Host: tcase.input,
				},
			}
			assert.Equal(t, tcase.expected, http_ctxtags.DefaultServiceNameDetector(req), "doesn't match")
		})

	}
}
