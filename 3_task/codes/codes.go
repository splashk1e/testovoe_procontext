package codes

import (
	"encoding/xml"
	"net/http"

	"golang.org/x/net/html/charset"
)

type ValCurses struct {
	Valutes []Valute `xml:"Valute"`
}

type Valute struct {
	ID   string `xml:"ID,attr"`
	Name string `xml:"Name"`
}

func GetValuteCodes() ValCurses {
	resp, err := http.Get("https://www.cbr-xml-daily.ru/daily.xml")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)

	decoder.CharsetReader = charset.NewReaderLabel

	var curses ValCurses
	if err := decoder.Decode(&curses); err != nil {
		panic(err)
	}

	return curses
}
