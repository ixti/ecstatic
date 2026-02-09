// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFetchMetadataTimeout(t *testing.T) {
	t.Run("returns default when env var is not set", func(t *testing.T) {
		assert := assert.New(t)

		result := getFetchMetadataTimeout()

		assert.Equal(defaultContainerMetadataTimeout, result)
	})

	t.Run("parses valid duration from env var", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4_TIMEOUT", "10s")

		result := getFetchMetadataTimeout()

		assert.Equal(10*time.Second, result)
	})

	t.Run("returns default when env var is invalid", func(t *testing.T) {
		assert := assert.New(t)

		t.Setenv("ECS_CONTAINER_METADATA_URI_V4_TIMEOUT", "not-a-duration")

		result := getFetchMetadataTimeout()

		assert.Equal(defaultContainerMetadataTimeout, result)
	})
}

func TestNewRootCommand(t *testing.T) {
	t.Run("has correct use and short description", func(t *testing.T) {
		assert := assert.New(t)

		cmd := NewRootCommand()

		assert.Equal("ecs-task-helper", cmd.Use)
		assert.Equal("ECS task helper utilities", cmd.Short)
	})

	t.Run("has metadata subcommand", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		cmd := NewRootCommand()

		metadataCmd, _, err := cmd.Find([]string{"metadata"})

		require.NoError(err)
		assert.Equal("metadata", metadataCmd.Use)
	})

	t.Run("has exec subcommand", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		cmd := NewRootCommand()

		execCmd, _, err := cmd.Find([]string{"exec"})

		require.NoError(err)
		assert.Contains(execCmd.Use, "exec")
	})

	t.Run("shows help without error", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		cmd := NewRootCommand()
		cmd.SetArgs([]string{"--help"})
		out := &bytes.Buffer{}
		cmd.SetOut(out)

		err := cmd.Execute()

		require.NoError(err)
		assert.Contains(out.String(), "ecs-task-helper")
		assert.Contains(out.String(), "metadata")
		assert.Contains(out.String(), "exec")
	})
}
