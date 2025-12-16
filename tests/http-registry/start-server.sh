#!/bin/bash
set -e
cd "$(dirname "$0")/server"
python3 -m http.server 8888 &
SERVER_PID=$!
echo $SERVER_PID > /tmp/http-registry-server.pid
echo "HTTP server started with PID $SERVER_PID"
sleep 2  # Wait for server to start


