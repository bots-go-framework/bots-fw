#!/usr/bin/env bash

rm $(pwd)/../.git/hooks/pre-commit
chmod +x git-hooks/pre-commit
ln -s $(pwd)/pre-commit $(pwd)/../.git/hooks/pre-commit
