// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: MIT

package container_metadata

import "errors"

var ErrMissingMetadataURI = errors.New("environment variable ECS_CONTAINER_METADATA_URI_V4 is missing")
