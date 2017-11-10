package httpwares_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/improbable-eng/go-httpwares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertingTripperware(t *testing.T, placeInChain int) httpwares.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		return httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for i := 0; i < placeInChain; i++ {
				require.NotEmpty(t, req.Header.Get(fmt.Sprintf("assert-%d", i)), "%d iteration of round tripper must have the headers from previous ones")
			}
			req.Header.Set(fmt.Sprintf("assert-%d", placeInChain), "true")
			return next.RoundTrip(req)
		})
	}
}

func TestTripperwareChainsInFifoOrder(t *testing.T) {
	numWares := 5
	retResp := &http.Response{StatusCode: 400}
	retErr := errors.New("some err")
	var wares []httpwares.Tripperware
	for i := 0; i < numWares; i++ {
		wares = append(wares, assertingTripperware(t, i))
	}
	finalTripper := httpwares.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		for i := 0; i < numWares; i++ {
			require.NotEmpty(t, req.Header.Get(fmt.Sprintf("assert-%d", i)), "the final round tripper must see the assserts of previous ones: %d", i)
		}
		return retResp, retErr
	})
	req, _ := http.NewRequest("GET", "http://whatever", nil)
	outResp, outErr := httpwares.WrapClient(&http.Client{Transport: finalTripper}, wares...).Transport.RoundTrip(req)
	assert.Equal(t, retResp, outResp)
	assert.Equal(t, retErr, outErr)
}
