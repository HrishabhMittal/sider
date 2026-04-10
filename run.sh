#!/bin/bash
go build -o /tmp/sider app/*.go
/tmp/sider "$@" &
SIDER_PID=$!
trap 'kill $SIDER_PID 2>/dev/null' EXIT
sleep 1
go run test/*.go
