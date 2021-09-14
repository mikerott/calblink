// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// TODO - make color fade from green to yellow to red
// TODO - add Clock type to manage time-of-day where we currently use Time and hacks to set it to the current day
// Configuration file:
// JSON file with the following structure:
// {
//   excludes: [ "event", "names", "to", "ignore"],
//   startTime: "hh:mm (24 hr format) to start blinking at every day",
//   endTime: "hh:mm (24 hr format) to stop blinking at every day",
//   skipDays: [ "weekdays", "to", "skip"],
//   calendar: "calendar"
//   responseState: "all"
//}
// Notes on items:
// Calendar is the calendar ID - the email address of the calendar.  For a person's calendar, that's their email.
// For a secondary calendar, it's the base64 string @group.calendar.google.com on the calendar details page.
// SkipDays may be localized.
// Excludes is exact string matches only.
// ResponseState can be one of: "all" (all events whatever their response status), "accepted" (only accepted events),
// "notRejected" (any events that are not rejected).  Default is notRejected.

// responseState is an enumerated list of event response states, used to control which events will activate the blink(1).
type responseState string

const (
	responseStateAll         = responseState("all")
	responseStateAccepted    = responseState("accepted")
	responseStateNotRejected = responseState("notRejected")
)

// checkStatus returns true if the given event status is one that should activate the blink(1) in the given responseState.
func (state responseState) checkStatus(status string) bool {
	switch state {
	case responseStateAll:
		return true

	case responseStateAccepted:
		return (status == "accepted")

	case responseStateNotRejected:
		return (status != "declined")
	}
	return false
}

func (state responseState) isValidState() bool {
	switch state {
	case responseStateAll:
		return true
	case responseStateAccepted:
		return true
	case responseStateNotRejected:
		return true
	}
	return false
}

// userPrefs is a struct that manages the user preferences as set by the config file and command line.

type userPrefs struct {
	excludes      map[string]bool
	startTime     *time.Time
	endTime       *time.Time
	skipDays      [7]bool
	calendar      string
	responseState responseState
}

// Struct used for decoding the JSON
type prefLayout struct {
	Excludes      []string
	StartTime     string
	EndTime       string
	SkipDays      []string
	Calendar      string
	ResponseState string
}

var (
	black               = "{\"color\":\"#000000\"}"
	green               = "{\"color\":\"#00FF00\"}"
	yellow              = "{\"color\":\"#FFFF00\"}"
	orange              = "{\"color\":\"#FFA500\"}"
	orangeFlashInfinite = "{\"pattern\":\"orange flash infinite\"}" // pattern must exist in the Blink(1) UI tool
	errorPattern        = "{\"pattern\":\"error\"}"                 // pattern must exist in the Blink(1) UI tool
)

// flags
var debugFlag = flag.Bool("debug", false, "Show debug messages")
var credentialsFlag = flag.String("credentials", "credentials.json", "Path to JSON file containing client secret, downloaded from Google calendar API enable")
var calNameFlag = flag.String("calendar", "primary", "Name of calendar to base blinker on (overrides value in config file)")
var configFileFlag = flag.String("config", "conf.json", "Path to configuration file")
var responseStateFlag = flag.String("response_state", "notRejected", "Which events to consider based on response: all, accepted, or notRejected")

var debugOut io.Writer = ioutil.Discard

// BEGIN GOOGLE CALENDAR API SAMPLE CODE

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-blink1.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// END GOOGLE CALENDAR API SAMPLE CODE

// Event viewing methods
func eventHasAcceptableResponse(item *calendar.Event, responseState responseState) bool {
	for _, attendee := range item.Attendees {
		if attendee.Self {
			return responseState.checkStatus(attendee.ResponseStatus)
		}
	}
	fmt.Fprintf(debugOut, "No self attendee found for %v\n", item)
	fmt.Fprintf(debugOut, "Attendees: %v\n", item.Attendees)
	return true
}

func nextEvent(items *calendar.Events, userPrefs *userPrefs) *calendar.Event {
	for _, i := range items.Items {
		if i.Start.DateTime != "" &&
			!userPrefs.excludes[i.Summary] &&
			eventHasAcceptableResponse(i, userPrefs.responseState) {
			return i
		}
	}
	return nil
}

func readUserPrefs() *userPrefs {
	userPrefs := &userPrefs{}
	// Set defaults from command line
	userPrefs.calendar = *calNameFlag
	userPrefs.responseState = responseState(*responseStateFlag)
	file, err := os.Open(*configFileFlag)
	defer file.Close()
	if err != nil {
		// Lack of a config file is not a fatal error.
		fmt.Fprintf(debugOut, "Unable to read config file %v : %v\n", *configFileFlag, err)
		return userPrefs
	}
	prefs := prefLayout{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&prefs)
	fmt.Fprintf(debugOut, "Decoded prefs: %v\n", prefs)
	if err != nil {
		log.Fatalf("Unable to parse config file %v", err)
	}
	if prefs.StartTime != "" {
		startTime, err := time.Parse("15:04", prefs.StartTime)
		if err != nil {
			log.Fatalf("Invalid start time %v : %v", prefs.StartTime, err)
		}
		userPrefs.startTime = &startTime
	}
	if prefs.EndTime != "" {
		endTime, err := time.Parse("15:04", prefs.EndTime)
		if err != nil {
			log.Fatalf("Invalid end time %v : %v", prefs.EndTime, err)
		}
		userPrefs.endTime = &endTime
	}
	userPrefs.excludes = make(map[string]bool)
	for _, item := range prefs.Excludes {
		fmt.Fprintf(debugOut, "Excluding item %v\n", item)
		userPrefs.excludes[item] = true
	}
	weekdays := make(map[string]int)
	for i := 0; i < 7; i++ {
		weekdays[time.Weekday(i).String()] = i
	}
	for _, day := range prefs.SkipDays {
		i, ok := weekdays[day]
		if ok {
			userPrefs.skipDays[i] = true
		} else {
			log.Fatalf("Invalid day in skipdays: %v", day)
		}
	}
	if prefs.Calendar != "" {
		userPrefs.calendar = prefs.Calendar
	}
	if prefs.ResponseState != "" {
		userPrefs.responseState = responseState(prefs.ResponseState)
		if !userPrefs.responseState.isValidState() {
			log.Fatalf("Invalid response state %v", prefs.ResponseState)
		}
	}
	fmt.Fprintf(debugOut, "User prefs: %v\n", userPrefs)
	return userPrefs
}

func tomorrow() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func setHourMinuteFromTime(t time.Time) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	flag.PrintDefaults()
}

func printStartInfo(userPrefs *userPrefs) {
	fmt.Fprintf(debugOut, "Checking calendar with ID %v\n", userPrefs.calendar)
	switch userPrefs.responseState {
	case responseStateAll:
		fmt.Fprintln(debugOut, "All events shown, regardless of accepted/rejected status.")
	case responseStateAccepted:
		fmt.Fprintln(debugOut, "Only accepted events shown.")
	case responseStateNotRejected:
		fmt.Fprintln(debugOut, "Rejected events not shown.")
	}
	if len(userPrefs.excludes) > 0 {
		fmt.Fprintln(debugOut, "Excluded events:")
		for item := range userPrefs.excludes {
			fmt.Fprintf(debugOut, "   %v\n", item)
		}
	}
	skipDays := ""
	join := ""
	for i, val := range userPrefs.skipDays {
		if val {
			skipDays += join
			skipDays += time.Weekday(i).String()
			join = ", "
		}
	}
	if len(skipDays) > 0 {
		fmt.Fprintln(debugOut, "Skip days: "+skipDays)
	}
	timeString := ""
	if userPrefs.startTime != nil {
		timeString += fmt.Sprintf("Time restrictions: after %02d:%02d", userPrefs.startTime.Hour(), userPrefs.startTime.Minute())
	}
	if userPrefs.endTime != nil {
		endTimeString := fmt.Sprintf("until %02d:%02d", userPrefs.endTime.Hour(), userPrefs.endTime.Minute())
		if len(timeString) > 0 {
			timeString += " and "
		} else {
			timeString += "Time restrictions: "
		}
		timeString += endTimeString
	}
	if len(timeString) > 0 {
		fmt.Fprintln(debugOut, timeString)
	}
}

func main() {

	flag.Usage = usage
	flag.Parse()

	if *debugFlag {
		debugOut = os.Stdout
	}

	userPrefs := readUserPrefs()

	// Overrides from command-line
	flag.Visit(func(myFlag *flag.Flag) {
		switch myFlag.Name {
		case "calendar":
			userPrefs.calendar = myFlag.Value.String()
		case "response_state":
			userPrefs.responseState = responseState(myFlag.Value.String())
			if !userPrefs.responseState.isValidState() {
				log.Fatalf("Invalid response state %v", userPrefs.responseState)
			}
		}
	})

	// BEGIN GOOGLE CALENDAR API SAMPLE CODE
	ctx := context.Background()

	b, err := ioutil.ReadFile(*credentialsFlag)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}
	// END GOOGLE CALENDAR API SAMPLE CODE

	printStartInfo(userPrefs)

	now := time.Now()
	weekday := now.Weekday()
	if userPrefs.skipDays[weekday] {
		tomorrow := tomorrow()
		untilTomorrow := tomorrow.Sub(now)
		fmt.Fprintf(debugOut, "Sleeping %v until tomorrow because it's a skip day\n", untilTomorrow)
		fmt.Fprint(os.Stdout, black)
		return
	}
	if userPrefs.startTime != nil {
		start := setHourMinuteFromTime(*userPrefs.startTime)
		fmt.Fprintf(debugOut, "Start time: %v\n", start)
		if diff := time.Since(start); diff < 0 {
			fmt.Fprintf(debugOut, "Sleeping %v because start time after now\n", -diff)
			fmt.Fprint(os.Stdout, black)
			return
		}
	}
	if userPrefs.endTime != nil {
		end := setHourMinuteFromTime(*userPrefs.endTime)
		fmt.Fprintf(debugOut, "End time: %v\n", end)
		if diff := time.Since(end); diff > 0 {
			tomorrow := tomorrow()
			untilTomorrow := tomorrow.Sub(now)
			fmt.Fprintf(debugOut, "Sleeping %v until tomorrow because end time %v before now\n", untilTomorrow, diff)
			fmt.Fprint(os.Stdout, black)
			return
		}
	}
	t := now.Format(time.RFC3339)
	events, err := srv.Events.List(userPrefs.calendar).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		// Leave the same color, set a flag. If we get more than a critical number of these,
		// set the color to blinking error pattern to tell the user we are in a failed state.
		fmt.Fprint(os.Stdout, errorPattern)
		return
	}

	next := nextEvent(events, userPrefs)
	var blinkState string
	if next != nil {
		startTime, err := time.Parse(time.RFC3339, next.Start.DateTime)
		if err == nil {
			delta := time.Since(startTime).Minutes()
			switch {
			case delta < -10: // more than 10 minutes until calendar event start
				blinkState = green
			case delta < -2: // more than 2, but less than 10 minutes until calendar event starts
				blinkState = yellow
			case delta < 0: // more than 0, but less than 2 minutes until calendar event starts
				blinkState = orangeFlashInfinite
			default: // in a meeting
				blinkState = orange
			}
			fmt.Fprint(os.Stdout, blinkState)
		} else {
			fmt.Fprintf(debugOut, "Error: %v\n", err)
			fmt.Fprint(os.Stdout, errorPattern)
		}
	}
}
