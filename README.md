# Blink(1) for Google Calendar (calblink)

## What is this?

Calblink is a small program to watch your Google Calendar and set a blink(1) USB
LED to change colors based on your next meeting. The colors it will use are:

*   Off: disconnected?
*   Green: more than 10 minutes until next calendar event
*   Yellow: less than 10, but more than 2 minutes to next calendar event
*   Flashing red: less than 2, but more than 0 minutes to next calendar event
*   Red: In meeting
*   Flashing red/blue: Error connecting to Calendar server.  This is to prevent
    the case where calblink silently fails and leaves you unaware that it has
    failed.

## What do I need use it?

To use calblink, you need the following:

1.  A blink(1) from [ThingM](http://blink1.thingm.com/) - calblink supports both
    mk1 and mk2 blink(1), but the mk2 is much nicer.
1.  A place to put the blink(1) where you can see it.
2.  The latest version of [Go](https://golang.org/).
3.  The calblink code, found in this directory.
4.  A directory to run this in.
5.  A few go packages, which we'll install later in the Setup section.
6.  A Google Calendar account.
7.  A Google Calendar OAuth 2 client ID. (We'll discuss getting one in the Setup
    section as well.)

## How do I set this up?

1.  Install Go, and plug your blink(1) in somewhere that you can see it.
2.  Bring up a command-line window, and create the directory you want to run
    this in. Set the GOPATH environment variable to point to this directory.
3.  Put calblink.go into the directory you just created.
4.  Install the Google APIs for Go:

    ```
    go get -u google.golang.org/api/calendar/v3
    go get -u golang.org/x/oauth2/...
    ```

5.  Get an OAuth 2 ID as described in step 1 of the [Google Calendar
    Quickstart](https://developers.google.com/google-apps/calendar/quickstart/go).
    Put the credentials.json file in your GOPATH directory.

6.  Run the calblink program: go run calblink.go

7.  It will request that you go to a URL and give it the token that you get
    back. You should access this URL from the account you want to read the
    calendar of.

8.  That's it! It should retrieve your calendar status and output the JSON
    that Blink1Control2 UI can parse.  Run it and see `go run calblink.go`

9.  Optionally set up a config file, as below.

10. Finally, set up Blink1Control2 to run the `blink1.sh` script on an
    interval of your choosing (but don't exhaust your Google Calendar API
    quota!).  Poke around in Blink1Control2 to set the UI to be backgrounded
    and auto-started, etc.

## What are the configuration options?

First off, run it with the --help option to see what the command-line options
are. Useful, perhaps, but maybe not what you want to use every time you run it.

calblink will look for a file named (by default) conf.json for its configuration
options. conf.json includes several useful options you can set:

*   excludes - a list of event titles which it will ignore. If you like blocking
    out time with "Make Time" or similar, you can add these names to the
    'excludes' array.
*   startTime - an HH:MM time (24-hour clock) which calblink won't turn on
    before. Because you might not want it turning on at 4am.
*   endTime - an HH:MM time (24-hour clock) which it won't turn on after.
*   skipDays - a list of days of the week that it should skip. A blink(1) in
    the offices doesn't need to run on Saturday/Sunday, after all, and if you
    WFH every Friday, why distract your coworkers?
*   calendar - which calendar to watch (defaults to primary). This is the email
    address of the calendar - either the calendar's owner, or the ID in its
    details page for a secondary calendar. "primary" is a magic string that
    means "the main calendar of the account whose auth token I'm using".
*   responseState - which response states are marked as being valid for a
    meeting. Can be set to "all", in which case any item on your calendar will
    light up; "accepted", in which case only items marked as 'accepted' on
    calendar will light up; or "notRejected", in which case items that you have
    rejected will not light up. Default is "notRejected".
*   deviceFailureRetries - how many times to retry accessing the blink(1) before
    failing out and terminating the program. Default is 10.

An example file:

```json
    {
        "excludes": ["Commute"],
        "skipDays": ["Saturday", "Sunday"],
        "startTime": "08:45",
        "endTime": "18:00",
        "calendar":"username@example.com",
        "responseState": "accepted"
    }
```

(Yes, the curly braces are required.)

## Known Issues

*   I have not done any special handling for Daylight Saving Time. There may
    be edge cases with sleeping around the DST change.
*   If there are more than 10 events that are skipped (all-day events, excluded
    events, and events with the wrong responseState) before the event that
    should be shown, the event will not be processed.

## Troubleshooting

*   If the blink(1) is flashing your error pattern, this means it was unable to
    connect to or authenticate to the Google Calendar server.  If your network is
    okay, your auth token may have expired.  Remove ~/.credentials/calendar-blink1.json
    and reconnect the app to your account.

## Legal

*   Calblink is not an official Google product.
*   Calblink is licensed under the Apache 2 license; see the LICENSE file for details.
*   Calblink contains code from the [Google Calendar API
    Quickstart](https://developers.google.com/google-apps/calendar/quickstart/go)
    which is licensed under the Apache 2 license.
*   All trademarks are the property of their respective holders.
