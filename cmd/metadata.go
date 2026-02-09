// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ixti/ecs-task-helper/pkg/container_metadata"
	"github.com/spf13/cobra"
)

type metadataCmdDeps struct {
	FetchMetadata func(ctx context.Context, timeout time.Duration) (*container_metadata.Metadata, error)
	Timeout       time.Duration
}

func defaultMetadataCmdDeps() *metadataCmdDeps {
	return &metadataCmdDeps{
		FetchMetadata: container_metadata.Fetch,
		Timeout:       getFetchMetadataTimeout(),
	}
}

func NewMetadataCommand(d *metadataCmdDeps) *cobra.Command {
	if d == nil {
		d = defaultMetadataCmdDeps()
	}

	format := "env"

	runE := func(cmd *cobra.Command, args []string) error {
		metadata, err := d.FetchMetadata(cmd.Context(), d.Timeout)

		if err != nil {
			if errors.Is(err, container_metadata.ErrMissingMetadataURI) {
				slog.Warn("Missing ECS metadata URI")

				if format == "json" {
					fmt.Fprintln(cmd.OutOrStdout(), "{}")
				}

				return nil
			}

			slog.Error("Can't retrieve ECS task metadata", "error", err)
			return err
		}

		switch format {
		case "json":
			data, _ := json.Marshal(metadata)
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
		case "env":
			for _, v := range metadata.Environ() {
				fmt.Fprintln(cmd.OutOrStdout(), v)
			}
		}

		return nil
	}

	cmd := &cobra.Command{
		Use:          "metadata",
		Short:        "Print ECS metadata as environment variables or JSON",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         runE,
	}

	cmd.Flags().StringVar(&format, "format", format, "Output format: env or json")

	return cmd
}
