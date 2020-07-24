#!/usr/bin/env bash
GOOS="linux" GOARCH="amd64" go build -o notifier
GOOS="windows" GOARCH="386" go build -o notifier_win

