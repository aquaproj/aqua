#!/bin/bash
if [ -f /tmp/http-registry-server.pid ]; then
  SERVER_PID=$(cat /tmp/http-registry-server.pid)
  echo "Stopping HTTP server with PID $SERVER_PID"
  kill $SERVER_PID 2>/dev/null || true
  rm /tmp/http-registry-server.pid
else
  echo "No HTTP server PID file found"
fi


