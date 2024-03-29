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
	h.bh.Handle(modules.LanguageMenu, th.CommandEqual("lang"))
	h.bh.Handle(modules.LanguageMenu, th.CallbackDataEqual("LanguageMenu"))
	h.bh.HandleCallbackQuery(modules.About, th.CallbackDataEqual("AboutMenu"))
	h.bh.HandleCallbackQuery(modules.Help, th.CallbackDataEqual("HelpMenu"))
	h.bh.Handle(modules.LanguageSet, th.CallbackDataPrefix("setLang"))
	h.bh.HandleMessage(modules.MediaDownloader, th.TextMatches(regexp.MustCompile(`(?:http(?:s)?://)?(?:www.|mobile.)?(?:twitter|x).com/.*?/([0-9]+)`)))
	h.bh.HandleMessage(modules.AccountInfo, th.CommandEqual("twitter"))
}
