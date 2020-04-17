#!/bin/sh

REQUIRED_ARGS=1
PROVIDED_ARGS=$#
VERSION=$1

check_args() {
  if [ "$REQUIRED_ARGS" != "$PROVIDED_ARGS" ];
  then
    echo "\nERROR: too few arguments. Please provide a version number.\n"
    echo ""
    echo "Example usage:"
    echo ""
    echo "$ build.sh 0.1.0"
    echo ""
    exit 1
  fi
  return 0
}

generate_asset() {
  if [[ -d bin ]]; then 
    rm -rf bin 
  fi
  mkdir bin 
  cp sensu-runbook bin/entrypoint
  tar -czf sensu-runbook-${VERSION}.tar.gz bin 
  rm -rf bin
}

check_args
generate_asset
