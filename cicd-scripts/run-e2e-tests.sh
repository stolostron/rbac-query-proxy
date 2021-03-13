#!/bin/bash
# Copyright (c) 2021 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

WORKDIR=`pwd`
git clone --depth 1 https://github.com/open-cluster-management/observability-e2e-test.git
cd observability-e2e-test

echo "Setup e2e test environment..."
./cicd-scripts/setup-e2e-tests.sh -a install
if [ $? -ne 0 ]; then
    echo "Cannot setup test environment"
    exit 1
fi

echo "Running e2e test.."
./cicd-scripts/run-e2e-tests.sh
if [ $? -ne 0 ]; then
    echo "Cannot pass all e2e test cases"
    exit 1
fi
