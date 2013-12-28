#!/bin/bash

# *************************************************************************************

# The MIT License (MIT)

# Copyright (c) 2013 Alex Edwards

# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

# *************************************************************************************

# Watch all *.go files in the specified directory
# Call the restart function when they are saved
function monitor() {
  inotifywait -q -m -r -e close_write --exclude '[^g][^o]$' $1 |
  while read line; do
    restart
  done
}

# Terminate and rerun the main Go program
function restart {
  if [ "$(pidof $PROCESS_NAME)" ]; then
    killall -q -w $PROCESS_NAME
  fi
  echo ">> Reloading..."
  go run $FILE_PATH $ARGS &
}

# Make sure all background processes get terminated
function close {
  killall -q -w inotifywait
  exit 0
}

trap close INT
echo "== Go-reload"
echo ">> Watching directories, CTRL+C to stop"

FILE_PATH=$1
FILE_NAME=$(basename $FILE_PATH)
PROCESS_NAME=${FILE_NAME%%.*}

shift
ARGS=$@

# Start the main Go program
go run $FILE_PATH $ARGS &

# Monitor the /src directories in all directories on the GOPATH
OIFS="$IFS"
IFS=':'
for path in $GOPATH
do
  monitor $path/src &
done
IFS="$OIFS"

# Monitor the current directory
monitor .
