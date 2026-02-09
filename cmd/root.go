// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const defaultContainerMetadataTimeout = 5 * time.Second

var version = "dev"

func getFetchMetadataTimeout() time.Duration {
	if v := os.Getenv("ECS_CONTAINER_METADATA_URI_V4_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}

		slog.Warn(
			"Invalid ECS_CONTAINER_METADATA_URI_V4_TIMEOUT, using default",
			"value", v,
			"default", defaultContainerMetadataTimeout,
		)
	}

	return defaultContainerMetadataTimeout
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ecs-task-helper",
		Short:   "ECS task helper utilities",
		Version: version,
	}

	cmd.AddCommand(NewMetadataCommand(nil))
	cmd.AddCommand(NewExecCommand(nil))
	cmd.AddCommand(NewCheckCommand(nil))

	return cmd
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
