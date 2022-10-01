#!/bin/bash

# to test, uncomment next 2 lines and wait a minute for calblink to run this script
# echo '{"color":"#446688"}'
# exit

# check if zoom is running
if ps aux | grep "/Applications/zoom.us.app/Contents/MacOS/zoom.us" | grep -v grep; then
  echo '{"color":"#FF0000"}'
  exit
fi

# Turn the light off during off hours
hour=$(date +%H)
if ((hour < 7 || hour >= 19)); then
  echo '{"color":"#000000"}'
  exit
fi

#!/bin/bash

today=$(date +%Y%m%d)
now=$(date +%H%M%S)
warning=300 # this is really 3 minutes, not 300 seconds

# check the google (really macOS) calendar
files=( $(find ~/Library/Calendars -name "*googlecom.ics" | xargs fgrep "DTSTART;TZID=America/Chicago:${today}T" | cut -d ":" -f1) )
for f in "${files[@]}"
do
  IFS=$'\n' lines=( $(sed -n '/^BEGIN:VEVENT/,/^END:VEVENT/p' $f) )
  for line in "${lines[@]}"
  do
    line=${line//[$'\t\r\n']}
    if [[ $line =~ ^SUMMARY:* ]]; then
      summary=$(echo "$line" | cut -c 9-)
    fi
    if [[ $line == DTSTART\;TZID=America/Chicago:${today}T* ]]; then
      dtstart=$(echo "$line" | rev | cut -d"T" -f1 | rev)
    fi
    if [[ $line == DTEND\;TZID=America/Chicago:${today}T* ]]; then
      dtend=$(echo "$line" | rev | cut -d"T" -f1 | rev)
    fi
    if [[ $line =~ ^.*mrheinheimer.*PARTSTAT.*$ ]]; then
      attendee=$(echo "$line" | rev | cut -d"=" -f1 | rev)
    fi
  done

  # all lines we care about have been read
  if [[ $attendee == "ACCE" ]] || [ -z $attendee ]; then # I accepted an invite or it's an event I made
    let diff=$((10#$dtstart))-$((10#$now))
    if [ $diff -lt $warning ] && [ $diff -gt 0 ]; then
      echo '{"color":"#FFFF00"}'
      exit
    elif [ $dtstart -lt $now ] && [ $now -lt $dtend ]; then
      echo '{"color":"#FF0000"}'
      exit
      echo "BLARG FILE: $i"
      echo "BLARG SUMMARY: $summary"
      echo "BLARG START: $dtstart"
      echo "BLARG END: $dtend"
      echo "BLARG STATUS: $attendee"
      echo ""
    fi
  fi
  unset summary
  unset dtstart
  unset dtend
  unset attendee
done

echo '{"color":"#00FF00"}'
  exit
