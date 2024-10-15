package bot

import (
	"github.com/yanzay/tbot/v2"
	"log"
	"strings"
)

func handleChatActionError(err error) {
	if err != nil {
		log.Println("Error chat action: ", err.Error())
	}
}

func handleMessageError(message *tbot.Message, err error) *tbot.Message {
	if err != nil {
		log.Printf("Message: %s\nError: %s", message.Text, err.Error())
	}
	return message
}

func replaceEmoji(emoji map[string]string, msg string) string {
	for placeholder, replacement := range emoji {
		msg = strings.ReplaceAll(msg, placeholder, replacement)
	}

	return msg
}
