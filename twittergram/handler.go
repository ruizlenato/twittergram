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
	h.bh.Use(database.SaveUsers)
	h.bh.Handle(modules.Start, th.CommandEqual("start"))
	h.bh.Handle(modules.Start, th.CallbackDataEqual("start"))
	h.bh.HandleCallbackQuery(modules.About, th.CallbackDataEqual("about"))
	h.bh.HandleMessage(modules.TwitterURL, th.TextMatches(regexp.MustCompile(`https?://(?:www.|mobile.)?(?:twitter|x).com/.*?/.*?/([0-9]+)`)))
}
