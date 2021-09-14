#!/bin/bash

# to test, uncomment next 2 lines and wait a minute for calblink to run this script
# echo '{"color":"#446688"}'
# exit

# check if zoom is running
if ps aux | grep "/Applications/zoom.us.app/Contents/MacOS/zoom.us" | grep -v grep; then
  echo '{"color":"#FF0000"}'
  exit
fi

# This is already done by conf.json
# don't call google calendar API if between 7pm and 8am
# hour=$(date +%H)
# if ((hour >= 19 || hour < 8)); then
#   echo '{"color":"#000000"}'
#   exit
# fi

# check the google calendar
export GOPATH=/Users/mrheinheimer/work/go
cd /Users/mrheinheimer/work/go/src/github.com/mikerott/calblink
if [ $? -ne 0 ]; then
  echo '{"pattern":"error"}'
  exit
fi
/usr/local/go/bin/go run calblink.go
