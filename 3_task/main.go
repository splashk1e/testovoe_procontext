package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/splashk1e/testovoe_procontext/codes"
)

type Answer struct {
	Name           string  `json:"name"`
	Min_price      float64 `json:"min_price"`
	Max_price      float64 `json:"max_price"`
	Avg_price      float64 `json:"avg_price"`
	Max_price_date string  `json:"max_price_date"`
	Min_price_date string  `json:"min_price_date"`
}

type Envelope struct {
	Body Body `xml:"Body"`
}

type Body struct {
	Response GetCursDynamicResponse `xml:"GetCursDynamicResponse"`
}

type GetCursDynamicResponse struct {
	Result GetCursDynamicResult `xml:"GetCursDynamicResult"`
}

type GetCursDynamicResult struct {
	Diffgram Diffgram `xml:"diffgram"`
}

type Diffgram struct {
	ValuteData ValuteData `xml:"ValuteData"`
}

type ValuteData struct {
	Currencies []ValuteCursDynamic `xml:"ValuteCursDynamic"`
}

type ValuteCursDynamic struct {
	CursDate string  `xml:"CursDate"`
	Vcurs    float64 `xml:"Vcurs"`
}

func main() {
	months := map[string]string{
		"January":   "01",
		"February":  "02",
		"March":     "03",
		"April":     "04",
		"May":       "05",
		"June":      "06",
		"July":      "07",
		"August":    "08",
		"September": "09",
		"October":   "10",
		"November":  "11",
		"December":  "12",
	}
	var answers []Answer
	url := "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"
	soapAction := "http://web.cbr.ru/GetCursDynamic"
	to_year, to_month, day := time.Now().Date()
	var to_day string
	var from_day string
	if day < 10 {
		to_day = "0" + strconv.Itoa(day)
	} else {
		to_day = strconv.Itoa(day)
	}
	from_year, from_month, day := time.Now().AddDate(0, 0, -90).Date()
	if day < 10 {
		from_day = "0" + strconv.Itoa(day)
	} else {
		from_day = strconv.Itoa(day)
	}

	toDate := fmt.Sprintf("%d-%s-%sT00:00:00", to_year, months[to_month.String()], to_day)
	fromDate := fmt.Sprintf("%d-%s-%sT00:00:00", from_year, months[from_month.String()], from_day)
	codes := codes.GetValuteCodes()
	for _, code := range codes.Valutes {
		soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
               xmlns:xsd="http://www.w3.org/2001/XMLSchema"
               xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetCursDynamic xmlns="http://web.cbr.ru/">
      <FromDate>%s</FromDate>
      <ToDate>%s</ToDate>
      <ValutaCode>%s</ValutaCode>
    </GetCursDynamic>
  </soap:Body>
</soap:Envelope>`, fromDate, toDate, code.ID)

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(soapBody))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("SOAPAction", soapAction)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		xmlStr := string(bodyBytes)
		//fmt.Println(xmlStr)
		var envelope Envelope
		if err := xml.Unmarshal([]byte(xmlStr), &envelope); err != nil {
			panic(err)
		}

		rates := envelope.Body.Response.Result.Diffgram.ValuteData.Currencies
		if len(rates) == 0 {
			fmt.Println("Нет данных для анализа.")
			return
		}
		min_date := rates[0].CursDate
		max_date := rates[0].CursDate
		min := rates[0].Vcurs
		max := rates[0].Vcurs
		averageCurs := 0.0
		for _, r := range rates {
			if r.Vcurs < min {
				min = r.Vcurs
				min_date = r.CursDate
			}
			if r.Vcurs > max {
				max = r.Vcurs
				max_date = r.CursDate
			}
			averageCurs += r.Vcurs
		}
		averageCurs /= float64(len(rates))
		fmt.Println(code.Name + ":")
		fmt.Printf("	Минимальный курс: %.4f\n", min)
		fmt.Printf("	Дата минимального курса: %s\n", min_date)
		fmt.Printf("	Максимальный курс: %.4f\n", max)
		fmt.Printf("	Дата максимального курса: %s\n", max_date)
		fmt.Printf("	Средний курс: %.4f\n", averageCurs)
		var answer Answer
		answer.Name = code.Name
		answer.Min_price = min
		answer.Max_price = max
		answer.Avg_price = averageCurs
		answer.Min_price_date = min_date
		answer.Max_price_date = max_date
		answers = append(answers, answer)
	}
	file, err := os.Create("answers.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Кодируем в JSON с отступами и пишем в файл
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(answers); err != nil {
		panic(err)
	}
}
