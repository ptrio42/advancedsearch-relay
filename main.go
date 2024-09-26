package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	relay := khatru.NewRelay()

	relay.Info.Name = "Advanced Search Relay"
	relay.Info.PubKey = "f1f9b0996d4ff1bf75e79e4cc8577c89eb633e68415c7faf74cf17a07bf80bd8"
	relay.Info.Description = "A relay to help you find your stuff on Nostr"
	relay.Info.Icon = ""

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		npub := strings.Trim(r.URL.Path, "/")

		fmt.Println("Pubkey", npub)

		if _, v, err := nip19.Decode(npub); err == nil {
			pubkey := v.(string)

			relay.QueryEvents = append(relay.QueryEvents, func(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
				filter.Authors = append(filter.Authors, pubkey)

				fmt.Println("Filter", filter)

				relay, err := nostr.RelayConnect(ctx, "wss://relay.nostr.band")
				if err != nil {
					panic(err)
				}

				before := `before:(\d{4})(-(\d{2}))?(-(\d{2}))?`
				beforeRegex := regexp.MustCompile(before)

				until := beforeRegex.FindStringSubmatch(filter.Search)
				fmt.Println("until length", len(until))

				if len(until) > 0 {
					fmt.Println("until", until)
					until := handleRange(until, filter.Search)
					filter.Until = &until
					filter.Search = beforeRegex.ReplaceAllString(filter.Search, "")
				}

				after := `after:(\d{4})(-(\d{2}))?(-(\d{2}))?`
				afterRegex := regexp.MustCompile(after)

				since := afterRegex.FindStringSubmatch(filter.Search)

				if len(since) > 0 {
					fmt.Println("since", since)
					since := handleRange(since, filter.Search)
					filter.Since = &since
					filter.Search = afterRegex.ReplaceAllString(filter.Search, "")
				}

				filter.Search = strings.TrimSpace(filter.Search)

				fmt.Println(filter)

				sub, err := relay.Subscribe(ctx, []nostr.Filter{filter})
				if err != nil {
					panic(err)
				}

				return sub.Events, nil
			})

		} else {
			panic(err)
		}

		relay.ServeHTTP(w, r)
	})

	fmt.Println("Listening on localhost:3366")

	http.ListenAndServe(":3366", nil)
}

func handleRange(matches []string, content string) nostr.Timestamp {
	year := matches[1]
	month := matches[3]
	day := matches[5]

	if month == "" {
		month = "01"
	}
	if day == "" {
		day = "01"
	}

	yearInt, _ := strconv.Atoi(year)
	monthInt, _ := monthFromString(month)
	dayInt, _ := strconv.Atoi(day)

	date := time.Date(yearInt, monthInt, dayInt, 0, 0, 0, 0, time.UTC)

	fmt.Println(date)
	return nostr.Timestamp(date.Unix())
}

func monthFromString(monthStr string) (time.Month, error) {
	monthInt, err := strconv.Atoi(monthStr)
	if err != nil || monthInt < 1 || monthInt > 12 {
		return 0, errors.New("invalid month value")
	}
	// Convert the integer to the Month type
	return time.Month(monthInt), nil
}
