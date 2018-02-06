#!/bin/bash
set -ex

CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -tags="containers_image_ostree_stub containers_image_openpgp exclude_graphdriver_devicemapper exclude_graphdriver_btrfs" cmd/executor/main.go
docker build -t gcr.io/priya-wadhwa/kbuilder:test .
