#!/usr/bin/env bash

# This script does not handle file names that contain spaces.

# TESTS
go test -v -race $(go list ./... | grep -v /vendor/)
RESULT=$?
[ $RESULT -ne 0 ] && exit 1

# FORMATTING
gofiles=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')
[ -z "$gofiles" ] && exit 0


unformatted=$(gofmt -l $gofiles)
[ -z "$unformatted" ] && exit 0

# Some files are not gofmt'd. Print message and fail.

echo >&2 "Go files must be formatted with gofmt. Please run:"
for fn in $unformatted; do
        echo >&2 "  gofmt -w $PWD/$fn"
done

exit 1
