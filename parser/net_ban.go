package parser

import (
	"math"
	"regexp"
	"time"
)

var (
	// [16:40:45][net_ban]: '123.123.123.124' banned for 120 minutes (test)

	// WE MUST match the whole line, otherwise players could exploit this regular expression by writing a specific chat line
	// matching this regular expression, which bans ANY or ALL ip ranges on all servers.
	// 0: whole match 1: IP 2: minutes 3: reason
	ddnetBanRegexp = regexp.MustCompile(`^(?i)\[[\d\- :.]+\]\[net_ban\]: banned '([\[\]:.0-9a-fA-F]+)' for (\d+) minutes? \((.*)\)$`)

	vanillaBanRegexp = regexp.MustCompile(`^(?i)\[[\d\- :.]+\]\[net_ban\]: '([\[\]:.0-9a-fA-F]+)' banned for (\d+) minutes? \((.*)\)$`)
)

func ParseClientBanned(line string) (cb ClientBanned, ok bool) {

	var (
		matches    []string
		ip         string
		minutesStr string
		reason     string
	)
	if matches = ddnetBanRegexp.FindStringSubmatch(line); len(matches) > 0 {
		ip = matches[1] // ipv6 requires [] around the ip
		minutesStr = matches[2]
		reason = matches[3]
	} else if matches = vanillaBanRegexp.FindStringSubmatch(line); len(matches) > 0 {
		ip = matches[1] // ipv6 requires [] around the ip
		minutesStr = matches[2]
		reason = matches[3]
	} else {
		return cb, false
	}

	return ClientBanned{
		IP:       ip,
		Duration: time.Duration(mustParseInt(minutesStr)) * time.Minute,
		Reason:   reason,
	}, true
}

type ClientBanned struct {
	IP       string        `json:"ip"`
	Duration time.Duration `json:"duration"`
	Reason   string        `json:"reason"`
}

func (cb ClientBanned) Minutes() int {
	return int(math.Ceil(cb.Duration.Minutes()))
}
