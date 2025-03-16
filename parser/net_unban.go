package parser

import (
	"regexp"
)

var (
	// WE MUST match the whole line, otherwise players could exploit this regular expression by writing a specific chat line
	// matching this regular expression, which would allow them to unban ips.
	// 0: all 1: IP
	// [2025-03-16 11:45:59][net_ban]: unbanned index 0 ('[36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06]')
	netUnbanIndexRegexp = regexp.MustCompile(`^(?i)\[[\d\- :.]+\]\[net_ban\]: unbanned index \d+ \('([\[\]:.0-9a-fA-F]+)'\)$`)

	// [2025-03-16 11:46:19][net_ban]: unbanned '[36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06]' for 1440 minutes (No reason given)
	// 0: all 1: IP 2: minutes 3: reason
	netUnbanIPRegexp = regexp.MustCompile(`^(?i)\[[\d\- :.]+\]\[net_ban\]: unbanned '([\[\]:.0-9a-fA-F]+)' for (\d+) minutes? \((.*)\)$`)
)

func ParseClientUnbanned(line string) (cb ClientUnbanned, ok bool) {

	var (
		matches []string
		ip      string
	)
	if matches = netUnbanIndexRegexp.FindStringSubmatch(line); len(matches) > 0 {
		ip = matches[1] // ipv6 requires [] around the ip
	} else if matches = netUnbanIPRegexp.FindStringSubmatch(line); len(matches) > 0 {
		ip = matches[1] // ipv6 requires [] around the ip
	} else {
		return cb, false
	}

	return ClientUnbanned{
		IP: ip,
	}, true
}

type ClientUnbanned struct {
	IP string `json:"ip"`
}
