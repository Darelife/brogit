#!/bin/bash

# Cleanup
rm -rf test_patching bin
pkill brogit-server || true

# Build
mkdir -p bin
go build -o bin/brogit-server ./cmd/brogit-server
go build -o bin/brogit-client ./cmd/brogit-client

# Setup Git Repo
mkdir test_patching
cd test_patching
git init
git config user.email "test@example.com"
git config user.name "Test Admin"
echo "Line 1: Hello" > hello.txt
echo "Line 2: World" >> hello.txt
echo "Line 3: Goodbye" >> hello.txt
git add hello.txt
git commit -m "Initial commit"

# Start Server
../bin/brogit-server &
SERVER_PID=$!
sleep 2

# Alice Edit 1: Change Line 1
# We need to simulate the file content properly.
# Alice sees the original file.
cp hello.txt alice_v1.txt
sed -i 's/Line 1: Hello/Line 1: Hello Alice/' alice_v1.txt

# Bob Edit 1: Change Line 3 (Based on original)
cp hello.txt bob_v1.txt
sed -i 's/Line 3: Goodbye/Line 3: Goodbye Bob/' bob_v1.txt

# Alice Edit 2: Change Line 2 (Based on her v1)
cp alice_v1.txt alice_v2.txt
sed -i 's/Line 2: World/Line 2: World Alice/' alice_v2.txt

# Push sequence: Alice1, Bob1, Alice2
echo "Pushing Alice V1..."
../bin/brogit-client push -user alice -file alice_v1.txt # Pushing as 'alice_v1.txt' but we want it to map to 'hello.txt'? 
# Ah, the client sends the filePath. The server uses that path.
# If I push 'alice_v1.txt', server thinks I'm editing 'alice_v1.txt'.
# I need to trick the client to push content but say it is 'hello.txt'.
# The client CLI takes '-file'. It reads content from that file AND sends that path.
# I need to modify client or just rename files locally before pushing.

# WORKAROUND: Rename files to hello.txt before pushing
cp alice_v1.txt hello.txt
../bin/brogit-client push -user alice -file hello.txt
sleep 1

cp bob_v1.txt hello.txt
../bin/brogit-client push -user bob -file hello.txt
sleep 1

cp alice_v2.txt hello.txt
../bin/brogit-client push -user alice -file hello.txt

# Trigger Commit
echo "Triggering commit..."
../bin/brogit-client commit > ../commit_output.txt

# Verify Content
echo "Final Content of hello.txt:"
cat hello.txt # This is just the local file Alice left. We need to check what Git has.
git checkout main # Force refresh from what server wrote? 
# Server writes to the directory. 'hello.txt' in the dir should be updated.
cat hello.txt

# Cleanup
kill $SERVER_PID
cd ..
# rm -rf test_patching
