package parser

import "regexp"

var (
	// id, target, nick, chat line
	// 2025-02-23 15:06:14 I chat: 0:-2:p'*mac: test
	ddnetChatLineRegexp = regexp.MustCompile(`chat: (\d+):(-?\d+):(.+): (.+)$`)

	// "[17:55:36][chat]: 0:-1:p'*mac: test"
	vanillaChatLineRegexp = regexp.MustCompile(`\[chat\]: (\d+):(-?\d+):(.+): (.+)$`)
)

type ChatMessage struct {
	ClientID int    `json:"client_id"`
	TargetID int    `json:"target_id"`
	Nickname string `json:"nickname"`
	Message  string `json:"message"`
}

func ParseChatMessage(line string) (cm ChatMessage, found bool) {

	var (
		clientID int
		targetID int
		nickname string
		message  string
	)
	if matches := ddnetChatLineRegexp.FindStringSubmatch(line); len(matches) > 0 {
		clientID = mustParseInt(matches[1])
		targetID = mustParseInt(matches[2])
		nickname = matches[3]
		message = matches[4]
	} else if matches := vanillaChatLineRegexp.FindStringSubmatch(line); len(matches) > 0 {
		clientID = mustParseInt(matches[1])
		targetID = mustParseInt(matches[2])
		nickname = matches[3]
		message = matches[4]
	} else {
		return cm, false
	}

	return ChatMessage{
		ClientID: clientID,
		TargetID: targetID,
		Nickname: nickname,
		Message:  message,
	}, true
}
