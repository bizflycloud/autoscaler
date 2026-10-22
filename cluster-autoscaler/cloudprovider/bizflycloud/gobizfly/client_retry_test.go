// This file is part of gobizfly

package gobizfly

import (
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud/gobizfly/testlib"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingTransport fails the n-th request to a path containing failPath with a network error.
type failingTransport struct {
	failPath string
	failOn   int32
	calls    int32
}

func (f *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, f.failPath) && atomic.AddInt32(&f.calls, 1) == f.failOn {
		return nil, errors.New("connection reset by peer")
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestDoRetryNetworkErrorReturnsError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(testlib.AuthURL(tokenPath), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"token": "xxx", "expires_at": "2019-11-22T15:39:54.000000Z"}`))
	})
	var l loadbalancer
	mux.HandleFunc(testlib.LoadBalancerURL(l.resourcePath()), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	// First call gets the 401, the retry after the token refresh fails at the network level
	client.httpClient = &http.Client{Transport: &failingTransport{failPath: l.resourcePath(), failOn: 2}}

	assert.NotPanics(t, func() {
		_, err := client.LoadBalancer.List(ctx, &ListOptions{})
		require.Error(t, err)
	})
}

func TestDoTokenEndpointUnauthorizedDoesNotRecurse(t *testing.T) {
	setup()
	defer teardown()

	var tokenCalls int32
	mux.HandleFunc(testlib.AuthURL(tokenPath), func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		w.WriteHeader(http.StatusUnauthorized)
	})
	var l loadbalancer
	mux.HandleFunc(testlib.LoadBalancerURL(l.resourcePath()), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.LoadBalancer.List(ctx, &ListOptions{})
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&tokenCalls))
}
