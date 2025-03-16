package parser

import "testing"

func TestParseChatLine(t *testing.T) {
	chat := `2024-11-25 01:12:00 I chat: 6:-2:scuf: b`

	msg, found := ParseChatMessage(chat)
	if !found {
		t.Fatalf("unexpected not found")
	}

	if msg.ClientID != 6 {
		t.Errorf("expected 6, got %d", msg.ClientID)
	}

	if msg.TargetID != -2 {
		t.Errorf("expected -2, got %d", msg.TargetID)
	}

	if msg.Nickname != "scuf" {
		t.Errorf("expected scuf, got %s", msg.Nickname)
	}

	if msg.Message != "b" {
		t.Errorf("expected b, got %s", msg.Message)
	}

}
