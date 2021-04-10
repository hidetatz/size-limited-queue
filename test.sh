#!/bin/bash

# make sure race condition is detected
output=$(go test -run Test_SingleThreadQueue_Push_Pop -race ./...)
if ! grep -q "WARNING: DATA RACE" <<< $output; then
    echo race condition is not detected in SingleThreadQueue test
    exit 1
fi

# make sure race condition is NOT detected
output=$(go test -run Test_MutexQueue_Push_Pop -race ./...)
if grep -q "WARNING: DATA RACE" <<< $output; then
    echo race condition is detected in MutexQueue test
    exit 1
fi

# make sure race condition is NOT detected
output=$(go test -run Test_SizeLimitedQueue_Push_Pop -race ./...)
if grep -q "WARNING: DATA RACE" <<< $output; then
    echo race condition is detected in SizeLimitedQueue test
    exit 1
fi
