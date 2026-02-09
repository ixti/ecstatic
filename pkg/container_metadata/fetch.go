// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type metadataPayload struct {
	ContainerARN   string `json:"ContainerARN"`
	ContainerName  string `json:"Name"`
	ContainerImage string `json:"Image"`
	Labels         struct {
		Cluster               string `json:"com.amazonaws.ecs.cluster"`
		TaskARN               string `json:"com.amazonaws.ecs.task-arn"`
		TaskDefinitionFamily  string `json:"com.amazonaws.ecs.task-definition-family"`
		TaskDefinitionVersion string `json:"com.amazonaws.ecs.task-definition-version"`
	} `json:"Labels"`
}

func Fetch(ctx context.Context, timeout time.Duration) (*Metadata, error) {
	// See: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint-v4-examples.html
	endpoint := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")

	if endpoint == "" {
		return nil, ErrMissingMetadataURI
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare metadata request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute metadata request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata request failed with status %d", res.StatusCode)
	}

	metadata := &metadataPayload{}
	if err := json.NewDecoder(res.Body).Decode(metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata response: %w", err)
	}

	return &Metadata{
		ContainerARN:          metadata.ContainerARN,
		ContainerName:         metadata.ContainerName,
		ContainerImage:        metadata.ContainerImage,
		TaskARN:               metadata.Labels.TaskARN,
		TaskDefinitionFamily:  metadata.Labels.TaskDefinitionFamily,
		TaskDefinitionVersion: metadata.Labels.TaskDefinitionVersion,
		ClusterName:           metadata.Labels.Cluster,
	}, nil
}
