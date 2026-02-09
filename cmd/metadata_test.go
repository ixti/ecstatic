// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ixti/ecs-task-helper/pkg/container_metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMetadata() *container_metadata.Metadata {
	return &container_metadata.Metadata{
		ContainerARN:          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		ContainerName:         "curl",
		ContainerImage:        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		TaskARN:               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		TaskDefinitionFamily:  "curltest",
		TaskDefinitionVersion: "24",
		ClusterName:           "default",
	}
}

func TestNewMetadataCommand(t *testing.T) {
	t.Run("with successful fetch outputs environ by default", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return testMetadata(), nil
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(out.String(), "ECS_CONTAINER_ARN=arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9\n")
		assert.Contains(out.String(), "ECS_CONTAINER_NAME=curl\n")
		assert.Contains(out.String(), "ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665\n")
		assert.Contains(out.String(), "ECS_TASK_DEFINITION_FAMILY=curltest\n")
		assert.Contains(out.String(), "ECS_TASK_DEFINITION_VERSION=24\n")
		assert.Contains(out.String(), "ECS_CLUSTER_NAME=default\n")
	})

	t.Run("with --format=env outputs environ", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return testMetadata(), nil
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		cmd.SetArgs([]string{"--format=env"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(out.String(), "ECS_CONTAINER_ARN=")
		assert.Contains(out.String(), "ECS_CONTAINER_NAME=curl\n")
	})

	t.Run("with --format=json outputs JSON", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return testMetadata(), nil
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		cmd.SetArgs([]string{"--format=json"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(out.String(), `"containerARN":"arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9"`)
		assert.Contains(out.String(), `"containerName":"curl"`)
		assert.Contains(out.String(), `"taskARN":"arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665"`)
		assert.Contains(out.String(), `"clusterName":"default"`)
	})

	t.Run("with missing metadata URI returns nil without error", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return nil, container_metadata.ErrMissingMetadataURI
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Empty(out.String())
	})

	t.Run("with missing metadata URI and --format=json outputs empty object", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return nil, container_metadata.ErrMissingMetadataURI
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		cmd.SetArgs([]string{"--format=json"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Equal("{}\n", out.String())
	})

	t.Run("with fetch error returns error", func(t *testing.T) {
		assert := assert.New(t)

		fetchErr := errors.New("network error")
		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return nil, fetchErr
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)

		err := cmd.Execute()

		assert.ErrorIs(err, fetchErr)
	})

	t.Run("passes timeout to fetch function", func(t *testing.T) {
		assert := assert.New(t)

		expectedTimeout := 10 * time.Second
		var receivedTimeout time.Duration

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				receivedTimeout = timeout
				return testMetadata(), nil
			},
			Timeout: expectedTimeout,
		}

		cmd := NewMetadataCommand(deps)
		cmd.SetOut(&bytes.Buffer{})

		cmd.Execute()

		assert.Equal(expectedTimeout, receivedTimeout)
	})

	t.Run("rejects positional arguments", func(t *testing.T) {
		assert := assert.New(t)

		deps := &metadataCmdDeps{
			FetchMetadata: func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error) {
				return testMetadata(), nil
			},
			Timeout: 5 * time.Second,
		}

		cmd := NewMetadataCommand(deps)
		cmd.SetArgs([]string{"unexpected-arg"})

		err := cmd.Execute()

		assert.Error(err)
	})

	t.Run("with nil deps uses defaults", func(t *testing.T) {
		cmd := NewMetadataCommand(nil)

		assert.Equal(t, "metadata", cmd.Use)
		assert.NotNil(t, cmd.RunE)
	})
}
