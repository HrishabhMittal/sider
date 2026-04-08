#!/bin/bash
go build -o /tmp/sider app/*.go
exec /tmp/sider "$@"
