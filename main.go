package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/telegram-bot-api"
	"log"

	"strconv"
)

type CallbackQueryPageData struct {
	Title string `json:"title"`
	Page  int    `json:"page"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)
	var GEO string

	for update := range updates {
		switch {
		case update.Message != nil:
			UserName := update.Message.From.UserName
			ChatID := update.Message.Chat.ID
			Text := update.Message.Text

			log.Printf("[%s] %d %s", UserName, ChatID, Text)
			log.Print(update.Message.Location)
			switch {
			case Text == "/start":
				msg := tgbotapi.NewMessage(ChatID, "Hello, "+update.Message.From.FirstName+"!")
				bot.Send(msg)
				msg = tgbotapi.NewMessage(ChatID, "Give me your location!")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonLocation("I'm here!")))
				bot.Send(msg)
				break
			case update.Message.Location != nil:
				log.Printf("%s", "User sent location")
				GEO = LocationToString(update.Message.Location)
				namesList, data := getPlaces(GEO)
				str, kb := PlacesInline(namesList, data, 0)
				msg := tgbotapi.NewMessage(ChatID, str)
				msg.ReplyMarkup = &kb
				bot.Send(msg)
				break
			default:
				log.Printf("%s", "Default")
				msg := tgbotapi.NewMessage(ChatID, "Give me your location!")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonLocation("I'm here!")))
				bot.Send(msg)
				break
			}
		case update.CallbackQuery != nil:
			var callBack CallbackQueryPageData
			json.Unmarshal([]byte(update.CallbackQuery.Data), &callBack)
			namesList, data := getPlaces(GEO)
			if callBack.Title == "places" {
				str, kb := PlacesInline(namesList, data, callBack.Page)
				msg := tgbotapi.NewEditMessageText(int64(update.CallbackQuery.From.ID), update.CallbackQuery.Message.MessageID, str)
				msg.ReplyMarkup = &kb
				bot.Send(msg)
			} else if callBack.Title == "showMap" {
				MAP := StringToLocation(data["coords"][callBack.Page])
				// Send NewLocation or NewVenue?
				msg := tgbotapi.NewVenue(int64(update.CallbackQuery.From.ID), namesList[strconv.Itoa(callBack.Page)], "", MAP.Latitude, MAP.Longitude)
				bot.Send(msg)			
				log.Printf("%s", "Map sent")

			}
		}
	}
}

func getPlaces(location string) (map[string]string, map[string][]string) {
	radius := 10
	response := getList(location, radius)
	for Len(response.Items[0].Item) == 0 {
		radius += 40
		response = getList(location, radius)
	}

	Places := make(map[string]string)
	for i, item := range response.Items[0].Item {
		Places[strconv.Itoa(i)] = HTML(item.Name[0].Text)
	}

	descs := GetReviews(response.Items[0].Item)
	pics := GetPhotoLinks(response.Items[0].Item)
	coords := GetCoordinates(response.Items[0].Item)
	data := make(map[string][]string)
	data["descs"] = descs
	data["pics"] = pics
	data["coords"] = coords
	return Places, data

}
func getList(coords string, radius int) APIResponse {
	newRequest := CreateRequestDependingOnRadius(radius, coords)
	xmlbody := xml.Header + string(newRequest)
	body := SendRequest(xmlbody)
	resp := GetResponse(body)

	return resp
}

func PlacesInline(Places map[string]string, data map[string][]string, page int) (string, tgbotapi.InlineKeyboardMarkup) {
	log.Printf("%s", "PlacesInline called")
	prevPage := 0
	nextPage := 0
	if page == 0 {
		prevPage = len(Places) - 1
		nextPage = page + 1
	}
	if page != 0 && page != len(Places) {
		prevPage = page - 1
		nextPage = page + 1
	}
	if page == len(Places)-1 {
		prevPage = page - 1
		nextPage = 0
	}

	description := HTML(data["descs"][page])
	if len(description) > MIN_COUNT_SYMBOL {
		data["descs"][page] = shortenDesc(description)
	}

	str := Places[strconv.Itoa(page)] + "\n"
	str += HTML(data["descs"][page]) + " \n"
	str += data["pics"][page]
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<", fmt.Sprintf("{ \"title\":\""+"places"+"\", \"page\":%d}", prevPage)),
			tgbotapi.NewInlineKeyboardButtonData("Map", fmt.Sprintf("{ \"title\":\""+"showMap"+"\", \"page\":%d}", page)),
			tgbotapi.NewInlineKeyboardButtonData(">", fmt.Sprintf("{ \"title\":\""+"places"+"\", \"page\":%d}", nextPage))))

	return str, kb
}
