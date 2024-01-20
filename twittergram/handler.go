package twittergram

import (
	"regexp"
	"twittergram/twittergram/database"
	"twittergram/twittergram/modules"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type Handler struct {
	bot *telego.Bot
	bh  *th.BotHandler
}

func NewHandler(bot *telego.Bot, bh *th.BotHandler) *Handler {
	return &Handler{
		bot: bot,
		bh:  bh,
	}
}

func (h *Handler) RegisterHandlers() {
	h.bh.HandleMessage(database.SaveUsers)
	h.bh.HandleMessage(modules.StartCmd, th.CommandEqual("start"))
}
