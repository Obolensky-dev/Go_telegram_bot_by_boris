
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type bResponse struct {
	Symbol string `json:"symbol"`
	Price float64 `json:"price,string""`
}

type wallet map[string]float64

var db = map[int]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("мой токен")
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		command := strings.Split(update.Message.Text, " ")
		userId := update.Message.From.ID

		switch command[0] {
		case "/start":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я тестовый бот Bitcoin кошелька от Бориса! Для помощи используй  /help"))
		case "/help":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Это твой личный тестовый Биткойн кошелёк!\nСписок команд (вводи без скобок []):\nДля добавления валюты - ADD [название валюты] [сумма (можно десятичное используй точку '.')]\nДля снятия валюты - SUB [название валюты] [сумма (можно десятичное используй точку '.')]\nДля удаления валюты - DEL [название валюты]\nДля демонстрации всех валют в кошельке - SHOW\nПримеры команд: \nADD BTC 0.15\nADD ETH 3.1225\nADD XRP 12.1\nSUB BTC 0.09\nDEL BTC\nSHOW"))

		case "ADD":
			if len(command)!= 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верно введена команда :( \n Для помощи используй  /help"))
				continue
			}

			_, err := getPriceUSD(command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная валюта! \n Для помощи используй  /help"))
				continue
			}

			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,err.Error()))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}

			db[userId][command[1]] += money

		case "SUB":
			if len(command)!= 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верно введена команда :( \n Для помощи используй  /help"))
				continue
			}

			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,err.Error()))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}

			db[userId][command[1]] -= money

		case "DEL":
			if len(command)!= 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не верно введена команда :( \n Для помощи используй  /help"))
				continue
			}

			delete(db[userId], command[1])

		case "SHOW":
			resp := ""
			for key, value := range db[userId]{
				usdPrice, err := getPriceUSD(key)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,err.Error()))
					continue
				}
				rubPrice, err := getPriceRUB()
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,err.Error()))
					continue
				}
				resp += fmt.Sprintf("%s: %.2f\nin USD: $%.2f\nin RUB: %.2f руб.\n\n",key, value, value * usdPrice, value * usdPrice* rubPrice)
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resp))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды :( \n Для помощи используй  /help"))

		}

	}
}

func getPriceUSD(symbol string) (float64, error) {

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT",symbol)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)
	if err != nil {
		return 0, err
	}

	if bRes.Symbol == "" {
		return 0, errors.New("Неверная валюта! \n Для помощи используй  /help")
	}

	return bRes.Price, nil
}

func getPriceRUB() (float64, error) {

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var bRes2 bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes2)
	if err != nil {
		return 0, err
	}

	return bRes2.Price, nil
}