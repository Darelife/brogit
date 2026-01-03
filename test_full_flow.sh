#!/bin/bash

# Cleanup
rm -rf test_repo bin
pkill brogit-server || true

# Build
mkdir -p bin
go build -o bin/brogit-server ./cmd/brogit-server
go build -o bin/brogit-client ./cmd/brogit-client

# Setup Git Repo
mkdir test_repo
cd test_repo
git init
git config user.email "prakharb2k6@gmail.com"
git config user.name "Darelife"
git checkout -b main
echo "Initial content" > initial.txt
git add initial.txt
git commit -m "Initial commit"

# Start Server (in the repo dir)
../bin/brogit-server &
SERVER_PID=$!
sleep 2

# Create some changes
echo "Alice was here" > alice.txt
../bin/brogit-client push -user alice -file alice.txt

sleep 1

echo "Bob was here" > bob.txt
../bin/brogit-client push -user bob -file bob.txt

# Trigger Commit
echo "Triggering commit..."
../bin/brogit-client commit > ../commit_output.txt

# Verify Git History
echo "Git Log:"
git log --graph --oneline --all

# Check branches
git branch -a

# Cleanup
kill $SERVER_PID
cd ..
# rm -rf test_repo
