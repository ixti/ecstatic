// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"

	"github.com/ixti/ecs-task-helper/pkg/container_metadata"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

type execCmdDeps struct {
	metadataCmdDeps
	Environ  func() []string
	LookPath func(file string) (string, error)
	Exec     func(argv0 string, argv []string, envv []string) error
}

func defaultExecCmdDeps() *execCmdDeps {
	return &execCmdDeps{
		metadataCmdDeps: *defaultMetadataCmdDeps(),
		Environ:    os.Environ,
		LookPath:   exec.LookPath,
		Exec:       unix.Exec,
	}
}

func NewExecCommand(d *execCmdDeps) *cobra.Command {
	if d == nil {
		d = defaultExecCmdDeps()
	}

	runE := func(cmd *cobra.Command, args []string) error {
		argv0, err := d.LookPath(args[0])
		if err != nil {
			slog.Error("Can't find command", "command", args[0], "error", err)
			return err
		}

		argv := append([]string{argv0}, args[1:]...)

		metadata, err := d.FetchMetadata(cmd.Context(), d.Timeout)

		var environ []string

		if err != nil {
			if !errors.Is(err, container_metadata.ErrMissingMetadataURI) {
				slog.Error("Can't retrieve ECS task metadata", "error", err)
				return err
			}

			slog.Warn("Missing ECS metadata URI, using current environment")
			environ = d.Environ()
		} else {
			environ = metadata.EnvironWith(d.Environ())
		}

		if err := d.Exec(argv0, argv, environ); err != nil {
			slog.Error("Command execution failed", "command", args[0], "error", err)
			return err
		}

		// This is effectively unreachable in real world, as Exec replaces the process.
		return nil
	}

	return &cobra.Command{
		Use:          "exec command [args...]",
		Short:        "Execute a command with ECS metadata environment variables",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE:         runE,
	}
}
