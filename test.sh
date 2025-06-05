#!/bin/bash

# Default number of iterations if not specified
iterations=${1:-50}

# Build the command
cmd="echo 'Hello, world!'"
for ((i=0; i<iterations; i++)); do
    cmd+=" | ./box -bc \"\$(tput setaf \$((RANDOM%7+1)))\""
done

# Run the command and measure time
time bash -c "$cmd"
