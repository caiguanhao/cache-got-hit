#!/bin/bash

set -e

REDIGO_COMMIT=4ed1111375cbeb698249ffe48dd463e9b0a63a7a
CLI_COMMIT=cf1f63a7274872768d4037305d572b70b1199397

rm -rf vendor
mkdir -p vendor

if test ! -d vendor/redigo-$REDIGO_COMMIT; then
  echo "getting redigo-$REDIGO_COMMIT..."
  curl -Ls https://github.com/garyburd/redigo/archive/$REDIGO_COMMIT.tar.gz | \
    tar -C vendor -xvzf -
fi

grep --include "*.go" -lrF "github.com/garyburd/redigo" . | \
  xargs sed -i "" "s,github.com/garyburd/redigo,github.com/caiguanhao/cache-got-hit/vendor/redigo-$REDIGO_COMMIT,"

if test ! -d vendor/cli-$CLI_COMMIT; then
  echo "getting cli-$CLI_COMMIT..."
  curl -Ls https://github.com/codegangsta/cli/archive/$CLI_COMMIT.tar.gz | \
    tar -C vendor -xvzf -
fi

grep --include "*.go" -lrF "github.com/codegangsta/cli" . | \
  xargs sed -i "" "s,github.com/codegangsta/cli,github.com/caiguanhao/cache-got-hit/vendor/cli-$CLI_COMMIT,"
