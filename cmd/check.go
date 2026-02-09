// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/spf13/cobra"
)

const defaultCheckTimeout = 1 * time.Second

type checkCmdDeps struct {
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
}

func defaultCheckCmdDeps() *checkCmdDeps {
	return &checkCmdDeps{
		HTTPClient: http.DefaultClient,
	}
}

func NewCheckCommand(d *checkCmdDeps) *cobra.Command {
	if d == nil {
		d = defaultCheckCmdDeps()
	}

	var (
		timeout  time.Duration
		statuses []int
		quiet    bool
	)

	runE := func(cmd *cobra.Command, args []string) error {
		url := args[0]

		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := d.HTTPClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if !quiet {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), string(body))
		}

		if !slices.Contains(statuses, resp.StatusCode) {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return nil
	}

	cmd := &cobra.Command{
		Use:          "check [flags] <url>",
		Short:        "Check HTTP endpoint availability",
		Long:         "Lightweight HTTP client for checking endpoint availability. Returns exit code 0 on success, 1 on failure.",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         runE,
	}

	cmd.Flags().DurationVar(&timeout, "timeout", defaultCheckTimeout, "Request timeout")
	cmd.Flags().IntSliceVar(&statuses, "status", []int{http.StatusOK}, "Expected HTTP status codes")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress output")

	return cmd
}
