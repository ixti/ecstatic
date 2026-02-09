// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewCheckCommand(t *testing.T) {
	t.Run("with successful request and matching status", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Equal("OK", out.String())
	})

	t.Run("with successful request and non-matching status", func(t *testing.T) {
		assert := assert.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusServiceUnavailable, "Service Unavailable"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		assert.Error(err)
		assert.Contains(err.Error(), "unexpected status code: 503")
		assert.Equal("Service Unavailable", out.String())
	})

	t.Run("with quiet flag suppresses output on success", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"--quiet", "http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Empty(out.String())
	})

	t.Run("with quiet flag suppresses output on failure", func(t *testing.T) {
		assert := assert.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusServiceUnavailable, "Service Unavailable"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"--quiet", "http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		assert.Error(err)
		assert.Empty(out.String())
	})

	t.Run("with custom status codes", func(t *testing.T) {
		require := require.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusNoContent, ""), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"--status", "200", "--status", "204", "http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
	})

	t.Run("with custom status codes using comma syntax", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusNoContent, ""), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"--status", "200,204", "http://localhost:8080/-/ready"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Empty(out.String())
	})

	t.Run("with HTTP client error", func(t *testing.T) {
		assert := assert.New(t)

		clientErr := errors.New("connection refused")
		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return nil, clientErr
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080/-/ready"})

		err := cmd.Execute()

		assert.ErrorIs(err, clientErr)
	})

	t.Run("with invalid URL", func(t *testing.T) {
		assert := assert.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"://invalid"})

		err := cmd.Execute()

		assert.Error(err)
	})

	t.Run("without URL argument", func(t *testing.T) {
		assert := assert.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)

		err := cmd.Execute()

		assert.Error(err)
	})

	t.Run("with too many arguments", func(t *testing.T) {
		assert := assert.New(t)

		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080", "http://localhost:8081"})

		err := cmd.Execute()

		assert.Error(err)
	})

	t.Run("passes correct URL to HTTP client", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var receivedURL string
		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					receivedURL = req.URL.String()
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080/-/ready"})
		cmd.SetOut(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(err)
		assert.Equal("http://localhost:8080/-/ready", receivedURL)
	})

	t.Run("uses GET method", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		var receivedMethod string
		deps := &checkCmdDeps{
			HTTPClient: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					receivedMethod = req.Method
					return mockResponse(http.StatusOK, "OK"), nil
				},
			},
		}

		cmd := NewCheckCommand(deps)
		cmd.SetArgs([]string{"http://localhost:8080/-/ready"})
		cmd.SetOut(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(err)
		assert.Equal(http.MethodGet, receivedMethod)
	})

	t.Run("with nil deps uses defaults", func(t *testing.T) {
		cmd := NewCheckCommand(nil)

		assert.Equal(t, "check [flags] <url>", cmd.Use)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("default timeout is 1 second", func(t *testing.T) {
		cmd := NewCheckCommand(nil)

		timeout, err := cmd.Flags().GetDuration("timeout")
		require.NoError(t, err)
		assert.Equal(t, defaultCheckTimeout, timeout)
	})

	t.Run("default status is 200", func(t *testing.T) {
		cmd := NewCheckCommand(nil)

		statuses, err := cmd.Flags().GetIntSlice("status")
		require.NoError(t, err)
		assert.Equal(t, []int{http.StatusOK}, statuses)
	})

	t.Run("default quiet is false", func(t *testing.T) {
		cmd := NewCheckCommand(nil)

		quiet, err := cmd.Flags().GetBool("quiet")
		require.NoError(t, err)
		assert.False(t, quiet)
	})
}
