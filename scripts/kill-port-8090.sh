#!/bin/bash

if lsof -i :8090 > /dev/null 2>&1; then
  kill $(lsof -t -i :8090)
  echo "Killed process on port 8090."
else
  echo "No process found on port 8090."
fi 