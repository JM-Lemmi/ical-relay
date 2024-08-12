#!/bin/bash

# args:
# -e: exit immediately on failure

success=1

for d in */; do
    n=${d#testcase-}; n=${n%/} # remove prefix to get test-id

    echo "[+] Test ${n} starting..."
    cd "${d}"

    if bash ./test.sh; then
        echo "[✔] Test ${n} passed"
    else
        echo "[✘] Test ${n} failed"
        success=0
        if [[ "${1}" == "-e" ]]; then
            exit 1
        fi
    fi
done

if [[ $success -eq 1 ]]; then
    exit 0
else
    exit 1
fi