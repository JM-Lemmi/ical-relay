#!/bin/bash

# args:
# -e: exit immediately on failure

success=1

# build common docker image in preparation
docker build -t ical-relay:test -f ../../cmd/ical-relay/Dockerfile ../../
docker build -t ical-notifier:test -f ../../cmd/ical-notifier/Dockerfile ../../

for d in */; do
    n=${d#testcase-}; n=${n%/} # remove prefix to get test-id

    echo "[+] Test ${n} starting..."
    cd "${d}"

    if bash ./test.sh; then
        echo "[✔] Test ${n} passed"
    else
        echo "[✘] Test ${n} failed"
        ./logs.sh
        success=0
        if [[ "${1}" == "-e" ]]; then
            ./cleanup.sh
            exit 1
        fi
    fi

    ./cleanup.sh
    cd ..
done

if [[ $success -eq 1 ]]; then
    exit 0
else
    exit 1
fi
