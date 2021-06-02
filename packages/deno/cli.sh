#!/bin/bash

deno run \
    --unstable \
    --allow-net \
    --allow-read \
    --allow-write \
    --allow-run \
    --allow-env \
    cli.ts $@
