#!/bin/bash

# Kill any running server
pkill brogit-server || true

# Build binaries
go build -o bin/brogit-server ./cmd/brogit-server
go build -o bin/brogit-client ./cmd/brogit-client

# Start server
./bin/brogit-server &
SERVER_PID=$!
sleep 2

# Create dummy file
echo "Hello Brogit" > test.txt

# Push changes
./bin/brogit-client push -user alice -file test.txt

# Wait a bit (simulate time passing)
sleep 1

# Push another change (simulate another user or same user)
echo "Hello Brogit v2" > test.txt
./bin/brogit-client push -user bob -file test.txt

# Commit
echo "Committing..."
./bin/brogit-client commit > commit_output.txt

# Check output
cat commit_output.txt

# Cleanup
kill $SERVER_PID
rm test.txt commit_output.txt bin/brogit-server bin/brogit-client
