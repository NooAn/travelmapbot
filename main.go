package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"github.com/telegram-bot-api"
)

type CallbackQueryPageData struct {
	Title string `json:"title"`
	Page  int    `json:"page"`
}

func main() {

	bot, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		LogPanic(err)
	}
	SetLogFile()
	bot.Debug = false
	Logf("Authorized on account %s", bot.Self.UserName)
	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)
	// ------------- 'global' variables used in main ----------
	var GEO string
	NumberOfFoundPlaces := 0
	// stuff we're getting from getAllPlaces
	var namesList map[string]string
	var data map[string][]string
	var types map[string][]string

	var ActiveSessionFlag bool 
	// --------------------------------------------------------

	for update := range updates {
		switch {
		case update.Message != nil:
			UserName := update.Message.From.UserName
			ChatID := update.Message.Chat.ID
			Text := update.Message.Text

			Logf("[%s] %d %s", UserName, ChatID, Text) //@TODO: check if username is empty, then print the ID? but can we find out who's this using the ID?
			Logf("User Location: %v", update.Message.Location)

			name := update.Message.From.FirstName
			if len(name) == 0 {
				name = "Путешественник"
			}

			integerText, textIsNotIntegerError := strconv.Atoi(Text)

			switch {
			case  textIsNotIntegerError == nil && ActiveSessionFlag:
				if integerText <= NumberOfFoundPlaces+1 && integerText > 0 {
					if integerText == NumberOfFoundPlaces+1 {
						str, kb := PlacesInline(namesList, data, 0)
						msg := tgbotapi.NewMessage(ChatID, str)
						msg.ReplyMarkup = &kb
						bot.Send(msg)
						ActiveSessionFlag = false
					} else {
						choice := data["types"][integerText-1]
						var chosenType string
						for i, t := range types["names"] {
							if t == choice{
								chosenType = types["IDs"][i]
							}
						}
						namesList, data = getChosenTypePlaces(GEO, chosenType)
						str, kb := PlacesInline(namesList, data, 0)
						msg := tgbotapi.NewMessage(ChatID, str)
						msg.ReplyMarkup = &kb
						bot.Send(msg)
						ActiveSessionFlag = false
					}
				} else {
					msg := tgbotapi.NewMessage(ChatID, "Попробуй еще раз. Нужно ввести просто число, без посторонних знаков.")
					bot.Send(msg)
				}
				break
			case Text == "/start":
				msg := tgbotapi.NewMessage(ChatID, "Привет, "+name+"!")
				bot.Send(msg)
				msg = tgbotapi.NewMessage(ChatID, "Чтобы поделиться своими координатами, нажми на кнопку \"посмотреть, что рядом!\".")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonLocation("Посмотреть, что рядом!")))
				bot.Send(msg)
				break
			case Text == "/help":
				Logf("/help " + UserName)
				msg := tgbotapi.NewMessage(ChatID, "Поделись своими координатами с мобильного устройства и я покажу тебе интересные места неподалеку. :)")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonLocation("Посмотреть, что рядом!")))
				bot.Send(msg)
			case update.Message.Location != nil:
				Logf("User %s sent location", update.Message.From.FirstName)
				GEO = LocationToString(update.Message.Location)
				namesList, data, types = getAllPlaces(GEO) // ODIN RAZ ETO DELAEM
				NumberOfFoundPlaces = len(data["types"])

				msg := tgbotapi.NewMessage(ChatID, "Вот, что я нашел недалеко от тебя. Что показать? (Введите только число)")  // social engineering huehue
				bot.Send(msg)
				str:= ListOfTypesToSend(data["types"])
				str = str + (strconv.Itoa(NumberOfFoundPlaces+1) + ".Показать всё!")
				msg = tgbotapi.NewMessage(ChatID, str)
				bot.Send(msg)

				ActiveSessionFlag = true
				break
			default:
				Log("Default, no action")
				msg := tgbotapi.NewMessage(ChatID, "Чтобы поделиться своими координатами, нажми на кнопку ниже.")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonLocation("Отправить координаты!")))
				bot.Send(msg)
				break
			}
		case update.CallbackQuery != nil:
			var callBack CallbackQueryPageData
			json.Unmarshal([]byte(update.CallbackQuery.Data), &callBack)
			if callBack.Title == "places" {
				str, kb := PlacesInline(namesList, data, callBack.Page)
				msg := tgbotapi.NewEditMessageText(int64(update.CallbackQuery.From.ID), update.CallbackQuery.Message.MessageID, str)
				msg.ReplyMarkup = &kb
				bot.Send(msg)
			} else if callBack.Title == "showMap" {
				MAP := StringToLocation(data["coords"][callBack.Page])
				msg2 := tgbotapi.NewVenue(int64(update.CallbackQuery.From.ID), namesList[strconv.Itoa(callBack.Page)], "", MAP.Latitude, MAP.Longitude)
				bot.Send(msg2)			
				log.Printf("%s", "Map sent")
			}
		}
	}
}

// this function is used for getting ALL the places that are around.
func getAllPlaces(location string) (map[string]string, map[string][]string, map[string][]string) {  // not sure if should NOT return this many args
	radius := 10
	response := GetListOfAllPlaces(location, radius)
	for Len(response.Items[0].Item) == 0 {
		radius += 40
		response = GetListOfAllPlaces(location, radius)
	}

	Places := make(map[string]string)
	for i, item := range response.Items[0].Item {
		Places[strconv.Itoa(i)] = HTML(item.Name[0].Text)
	}
	fmt.Println("PLACES: ", Places)
	descs := GetReviews(response.Items[0].Item)
	pics := GetPhotoLinks(response.Items[0].Item)
	coords := GetCoordinates(response.Items[0].Item)
	
	typeIDsWeHave := GetTypes(response.Items[0].Item)
	typeNames := make(map[string][]string)
	typeNames = GetTypeNames(GetListOfTypes().Items[0].Item)

	// пtranslating to human language
	var allTypesWeHave []string
	for _, typeID := range typeIDsWeHave {
		for i, id := range typeNames["IDs"] {
			if (typeID==id) {
				allTypesWeHave = append(allTypesWeHave, typeNames["names"][i])
			}
		}
	}

	allTypesWeHaveSet := makeSet(allTypesWeHave)

	data := make(map[string][]string)
	data["descs"] = descs
	data["pics"] = pics
	data["coords"] = coords
	var destins []string
	for _, geo := range coords {
		destins = append(destins, calculateDistance(location, geo))
	}
	data["destins"] = destins
	data["types"] = allTypesWeHaveSet
	return Places, data, typeNames 
}

func getChosenTypePlaces(location string, usrType string) (map[string]string, map[string][]string) {
	radius := 10
	response := GetListOfChosenTypePlaces(location, radius, usrType)
	for Len(response.Items[0].Item) == 0 {
		radius += 40
		response = GetListOfChosenTypePlaces(location, radius, usrType)
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
	var destins []string
	for _, geo := range coords {
		destins = append(destins, calculateDistance(location, geo))
	}
	data["destins"] = destins
	return Places, data
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
	if len(description) > MAX_LENGTH {
		data["descs"][page] = shortenDesc(description)
	}

	str := "Место " + strconv.Itoa(page+1) + " из " + strconv.Itoa(len(Places)) + ": \n"
	str += Places[strconv.Itoa(page)] + "\n"
	str += HTML(data["descs"][page]) + " \n"
	str += "\nНа расстоянии " + data["destins"][page] + " км" + "\n"
	str += data["pics"][page]
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("<", fmt.Sprintf("{ \"title\":\""+"places"+"\", \"page\":%d}", prevPage)),
			tgbotapi.NewInlineKeyboardButtonData("Карта", fmt.Sprintf("{ \"title\":\""+"showMap"+"\", \"page\":%d}", page)),
			tgbotapi.NewInlineKeyboardButtonData(">", fmt.Sprintf("{ \"title\":\""+"places"+"\", \"page\":%d}", nextPage))))

	return str, kb
}

func ListOfTypesToSend(types []string) string {  // Still not sure about the method's name
	var str string
	for i, t := range types {
		str += strconv.Itoa(i+1) + "." + t + "\n"
	}

	return str
}

func makeSet(listOfElements []string) []string {
	Set := make(map[string]bool)
	Set[listOfElements[0]] = true
	for _, x := range listOfElements {
		if !(Set[x]){
			Set[x] = true
		}
	}
	finalSet := make([]string, 0, len(Set))
	for name := range Set {
		finalSet = append(finalSet, name)
	} 
	return finalSet
}