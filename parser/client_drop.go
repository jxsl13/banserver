package parser

import "regexp"

var (
	// 0: full 1: ID 2: IP 3: port
	// [server]: client dropped. cid=0 addr=193.91.82.245:57603 reason=''
	vanillaDroppedRegex = regexp.MustCompile(`\[server\]: client dropped\. cid=(\d+) addr=([\[\]:.0-9a-fA-F]+):(\d+) reason='(.*)'`)


)

type ClientDropped struct {
	ClientID int    `json:"client_id"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Reason   string `json:"reason"`
}

func ParseClientDropped(line string) (_ ClientDropped, ok bool) {

	var (
		matches []string
		idStr   string
		ip      string
		portStr string
		reason  string
	)
	if matches = vanillaDroppedRegex.FindStringSubmatch(line); len(matches) > 0 {
		idStr = matches[1]
		ip = matches[2]
		portStr = matches[3]
		reason = matches[4]
	} else {
		return ClientDropped{}, false
	}

	return ClientDropped{
		ClientID: mustParseInt(idStr),
		IP:       ip,
		Port:     mustParseInt(portStr),
		Reason:   reason,
	}, true
}
