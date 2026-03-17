// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMetadataWithClusterName() *Metadata {
	return &Metadata{
		ContainerARN:          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		ContainerName:         "curl",
		ContainerImage:        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		TaskARN:               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		TaskDefinitionFamily:  "curltest",
		TaskDefinitionVersion: "24",
		ClusterName:           "default",
	}
}

func testMetadataWithClusterARN() *Metadata {
	return &Metadata{
		ContainerARN:          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		ContainerName:         "curl",
		ContainerImage:        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		TaskARN:               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		TaskDefinitionFamily:  "curltest",
		TaskDefinitionVersion: "24",
		ClusterARN:            "arn:aws:ecs:us-west-2:111122223333:cluster/default",
		ClusterName:           "default",
	}
}

func expectedOverridesWithClusterName() []string {
	return []string{
		"ECS_CONTAINER_ARN=arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		"ECS_CONTAINER_NAME=curl",
		"ECS_CONTAINER_IMAGE=111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		"ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		"ECS_TASK_ID=8f03e41243824aea923aca126495f665",
		"ECS_TASK_DEFINITION_FAMILY=curltest",
		"ECS_TASK_DEFINITION_VERSION=24",
		"ECS_CLUSTER_ARN=",
		"ECS_CLUSTER_NAME=default",
	}
}

func expectedOverridesWithClusterARN() []string {
	return []string{
		"ECS_CONTAINER_ARN=arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
		"ECS_CONTAINER_NAME=curl",
		"ECS_CONTAINER_IMAGE=111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
		"ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		"ECS_TASK_ID=8f03e41243824aea923aca126495f665",
		"ECS_TASK_DEFINITION_FAMILY=curltest",
		"ECS_TASK_DEFINITION_VERSION=24",
		"ECS_CLUSTER_ARN=arn:aws:ecs:us-west-2:111122223333:cluster/default",
		"ECS_CLUSTER_NAME=default",
	}
}

func TestMetadata_TaskID(t *testing.T) {
	t.Run("with cluster name", func(t *testing.T) {
		metadata := &Metadata{
			ClusterName: "default",
			TaskARN:     "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "8f03e41243824aea923aca126495f665", metadata.TaskID())
	})

	t.Run("with mismatched cluster name uses fallback", func(t *testing.T) {
		metadata := &Metadata{
			ClusterName: "other-cluster",
			TaskARN:     "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "8f03e41243824aea923aca126495f665", metadata.TaskID())
	})

	t.Run("with empty cluster name uses fallback", func(t *testing.T) {
		metadata := &Metadata{
			TaskARN: "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "8f03e41243824aea923aca126495f665", metadata.TaskID())
	})

	t.Run("with legacy task ARN format", func(t *testing.T) {
		metadata := &Metadata{
			ClusterName: "default",
			TaskARN:     "arn:aws:ecs:us-west-2:111122223333:task/8f03e41243824aea923aca126495f665",
		}

		assert.Equal(t, "8f03e41243824aea923aca126495f665", metadata.TaskID())
	})

	t.Run("with blank TaskARN", func(t *testing.T) {
		metadata := &Metadata{}
		assert.Equal(t, "", metadata.TaskID())
	})
}

func TestMetadata_ToJSON(t *testing.T) {
	t.Run("without cluster ARN", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		data, err := json.Marshal(testMetadataWithClusterName())
		require.NoError(err)

		var result map[string]string
		require.NoError(json.Unmarshal(data, &result))

		assert.Equal(map[string]string{
			"containerARN":          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
			"containerName":         "curl",
			"containerImage":        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
			"taskARN":               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
			"taskDefinitionFamily":  "curltest",
			"taskDefinitionVersion": "24",
			"clusterName":           "default",
		}, result)
	})

	t.Run("with cluster ARN", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		data, err := json.Marshal(testMetadataWithClusterARN())
		require.NoError(err)

		var result map[string]string
		require.NoError(json.Unmarshal(data, &result))

		assert.Equal(map[string]string{
			"containerARN":          "arn:aws:ecs:us-west-2:111122223333:container/0206b271-b33f-47ab-86c6-a0ba208a70a9",
			"containerName":         "curl",
			"containerImage":        "111122223333.dkr.ecr.us-west-2.amazonaws.com/curltest:latest",
			"taskARN":               "arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665",
			"taskDefinitionFamily":  "curltest",
			"taskDefinitionVersion": "24",
			"clusterARN":            "arn:aws:ecs:us-west-2:111122223333:cluster/default",
			"clusterName":           "default",
		}, result)
	})
}

func TestMetadata_Environ(t *testing.T) {
	t.Run("without cluster ARN", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(expectedOverridesWithClusterName(), testMetadataWithClusterName().Environ())
	})

	t.Run("with cluster ARN", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(expectedOverridesWithClusterARN(), testMetadataWithClusterARN().Environ())
	})
}

func TestMetadata_EnvironWith(t *testing.T) {
	t.Run("with nil base returns only overrides", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expectedOverridesWithClusterName(), testMetadataWithClusterName().EnvironWith(nil))
	})

	t.Run("with empty base returns only overrides", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(expectedOverridesWithClusterName(), testMetadataWithClusterName().EnvironWith([]string{}))
	})

	t.Run("replaces existing metadata env vars", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{
			"ECS_CONTAINER_NAME=old-value",
			"ECS_TASK_ARN=old-task-arn",
			"ECS_CLUSTER_ARN=old-cluster-arn",
			"PATH=/usr/bin",
		}

		env := testMetadataWithClusterARN().EnvironWith(base)

		assert.NotContains(env, "ECS_CONTAINER_NAME=old-value")
		assert.NotContains(env, "ECS_TASK_ARN=old-task-arn")
		assert.NotContains(env, "ECS_CLUSTER_ARN=old-cluster-arn")
		assert.Contains(env, "ECS_CONTAINER_NAME=curl")
		assert.Contains(env, "ECS_TASK_ARN=arn:aws:ecs:us-west-2:111122223333:task/default/8f03e41243824aea923aca126495f665")
		assert.Contains(env, "ECS_CLUSTER_ARN=arn:aws:ecs:us-west-2:111122223333:cluster/default")
		assert.Contains(env, "PATH=/usr/bin")
	})

	t.Run("preserves unrelated env vars", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{
			"PATH=/usr/bin",
			"HOME=/home/test",
			"CUSTOM_VAR=custom-value",
			"ECS_SOME_OTHER_VAR=should-remain",
		}

		env := testMetadataWithClusterName().EnvironWith(base)

		assert.Contains(env, "PATH=/usr/bin")
		assert.Contains(env, "HOME=/home/test")
		assert.Contains(env, "CUSTOM_VAR=custom-value")
		assert.Contains(env, "ECS_SOME_OTHER_VAR=should-remain")
	})

	t.Run("overrides appear at end", func(t *testing.T) {
		assert := assert.New(t)

		base := []string{"PATH=/usr/bin", "HOME=/home/test"}
		env := testMetadataWithClusterName().EnvironWith(base)

		overrides := expectedOverridesWithClusterName()
		envLen := len(env)
		overridesLen := len(overrides)

		assert.Equal(envLen, len(base)+overridesLen)
		assert.Equal(overrides, env[envLen-overridesLen:])
	})
}
