package main

import (
	"github.com/telegram-bot-api"
	"log"
	"os"
	"fmt"
)
type TBWrap struct {
	bot     *tgbotapi.BotAPI
	running bool
}


func (margelet *TBWrap) Run() error {
	Log("bot run logger")

	if (margelet.bot == nil) {
		Log("bot nil")
		bot, err := tgbotapi.NewBotAPI(TOKEN_BOTLOGER)
		if err != nil {
			log.Panic(err)
		}
		margelet.bot = bot
		margelet.running = true
	}
	updates, err := margelet.bot.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})

	if err != nil {
		Log(err.Error())
		return err
	}

	for margelet.running {
		select {
		case update := <-updates:
			message := update.Message
			ChatID := update.Message.Chat.ID
			fmt.Println(ChatID)
			fmt.Println(message)
		}
	}
	return nil
}

func (margelet *TBWrap) Stop() {
	margelet.running = false
}

func (c *TBWrap) Send(currentChatId int64, msg string) {
	tgmsg := tgbotapi.NewMessage(currentChatId, msg)
	if (c.bot == nil) {
		fmt.Println("Bot logger start....")
		bot, err := tgbotapi.NewBotAPI(TOKEN_BOTLOGER)
		if err != nil {
			log.Panic(err)
		}
		c.bot = bot
		Log("bot started")
	}
	c.bot.Send(tgmsg)
}
func SetLogFile() {
	f, err := os.Create("/log/log_bot.txt")
	if err != nil {
		CheckErr(err,"logfile open failed")
	} else {
		log.SetOutput(f)
	}
}