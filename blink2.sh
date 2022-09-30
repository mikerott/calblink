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

today=$(date +%Y%m%d)
now=$(date +%H%M%S)

# check the google (really macOS) calendar
files=( $(find ~/Library/Calendars -name "*googlecom.ics" | xargs fgrep "DTSTART;TZID=America/Chicago:${today}T" | cut -d ":" -f1) )
for i in "${files[@]}"
do
  while IFS="" read -r p || [ -n "$p" ]
  do
    if [[ $p =~ ^SUMMARY:* ]]; then
      summary=$(echo "$p" | cut -c 9-)
    fi
    if [[ $p == DTSTART\;TZID=America/Chicago:${today}T* ]]; then
      dtstart=$(echo "$p" | rev | cut -d"T" -f1 | rev)
    fi
    if [[ $p == DTEND\;TZID=America/Chicago:${today}T* ]]; then
      dtend=$(echo "$p" | rev | cut -d"T" -f1 | rev)
    fi
    if [[ $p =~ ^.*mrheinheimer.*PARTSTAT.*$ ]]; then
      attendee=$(echo "$p" | rev | cut -d"=" -f1 | rev)
    fi
    if [ -n "$dtstart" ] && [ -n "$dtend" ] && [ -n "$summary" ] && [ -n "$attendee" ]; then
      # we've collected all the info we need for this segment
      summary=${summary//[$'\t\r\n']}
      dtstart=${dtstart//[$'\t\r\n']}
      dtend=${dtend//[$'\t\r\n']}
      attendee=${attendee//[$'\t\r\n']}
      if [ $attendee == "ACCE" ] && [ ($dtstart - 3) -lt $now ]; then
        echo '{"color":"#FFFF00"}'
        exit
      elif [ $attendee == "ACCE" ] && [ $dtstart -lt $now ] && [ $now -lt $dtend ]; then
        echo '{"color":"#FF0000"}'
        exit
        #echo "BLARG FILE: $i"
        #echo "BLARG SUMMARY: $summary"
        #echo "BLARG START: $dtstart"
        #echo "BLARG END: $dtend"
        #echo "BLARG STATUS: $attendee"
        #echo ""
      fi
      segmentdone=true
    fi
    if [[ $p =~ ^BEGIN.*$ ]] || [ -n "$segmentdone" ]; then
      unset summary
      unset dtstart
      unset dtend
      unset attendee
      unset segmentdone
    fi
  done < $i
done

echo '{"color":"#00FF00"}'
  exit
