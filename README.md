# Blink(1) for Google Calendar (calblink)

## What is this?

Calblink is a small program to watch your Google Calendar and set a blink(1) USB
LED to change colors based on your next meeting. The colors it will use are:

## What do I need use it?

1. Configure MacOS native Calendar app to sync with your Google calendar, but not
   too frequently, so you don't use up your Google Calendar API quota.

2. Grant Blink1Control2 "Full Disk Access" in the Privacy tab of Security & Privacy system prefs

3. Set up Blink1Control2 to run the blink2.sh script (which reads files on the disk)  at whatever frequency you like.

## Legal

*   Calblink is not an official Google product.
*   Calblink is licensed under the Apache 2 license; see the LICENSE file for details.
*   Calblink contains code from the [Google Calendar API
    Quickstart](https://developers.google.com/google-apps/calendar/quickstart/go)
    which is licensed under the Apache 2 license.
*   All trademarks are the property of their respective holders.
