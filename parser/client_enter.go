package parser

import (
	"regexp"
)

var (
	// 0: full 1: ID 2: IP 3: port
	gameEnterRegex = regexp.MustCompile(`(?i)player has entered the game\. ClientID=([\d]+) addr=[<{]{0,2}([\[\]:.0-9a-fA-F]+):(\d+)[}>]{0,2}`)
)

func ParseClientEntered(line string) (_ ClientEntered, ok bool) {

	var (
		matches   []string
		joinIDStr string
		joinIP    string
		portStr   string
	)
	if matches = gameEnterRegex.FindStringSubmatch(line); len(matches) > 0 {
		joinIDStr = matches[1]
		joinIP = matches[2]
		portStr = matches[3]
	} else {
		return ClientEntered{}, false
	}

	return ClientEntered{
		ClientID: mustParseInt(joinIDStr),
		IP:       joinIP,
		Port:     mustParseInt(portStr),
	}, true
}

type ClientEntered struct {
	ClientID int    `json:"client_id"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}
