// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	t.Run("with missing env var", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", "")

		metadata, err := Fetch(context.Background(), 5*time.Second)

		assert.Nil(metadata)
		assert.ErrorIs(err, ErrMissingMetadataURI)
	})

	t.Run("with successful response", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"ContainerARN": "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
				"Name": "curl",
				"Image": "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
				"Labels": {
					"com.amazonaws.ecs.cluster": "default",
					"com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
					"com.amazonaws.ecs.task-definition-family": "curltest",
					"com.amazonaws.ecs.task-definition-version": "24"
				}
			}`))
		}))
		defer server.Close()

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", server.URL)

		metadata, err := Fetch(context.Background(), 5*time.Second)

		require.NoError(err)

		assert.Equal(&Metadata{
			ContainerARN:          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
			ContainerName:         "curl",
			ContainerImage:        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
			TaskARN:               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
			TaskDefinitionFamily:  "curltest",
			TaskDefinitionVersion: "24",
			ClusterName:           "default",
		}, metadata)
	})

	t.Run("with non-OK status", func(t *testing.T) {
		assert := assert.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", server.URL)

		metadata, err := Fetch(context.Background(), 5*time.Second)

		assert.Nil(metadata)
		assert.ErrorContains(err, "metadata request failed with status 500")
	})

	t.Run("with invalid JSON", func(t *testing.T) {
		assert := assert.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", server.URL)

		metadata, err := Fetch(context.Background(), 5*time.Second)

		assert.Nil(metadata)
		assert.ErrorContains(err, "failed to decode metadata response")
	})

	t.Run("with canceled context", func(t *testing.T) {
		assert := assert.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		metadata, err := Fetch(ctx, 5*time.Second)

		assert.Nil(metadata)
		assert.ErrorContains(err, "failed to execute metadata request")
	})

	t.Run("with timeout", func(t *testing.T) {
		assert := assert.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4", server.URL)

		metadata, err := Fetch(context.Background(), 10*time.Millisecond)

		assert.Nil(metadata)
		assert.ErrorContains(err, "failed to execute metadata request")
	})
}
