package mdn

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func GetHttp(httpCode int) (map[int]string, error) {

	mapDescription := make(map[int]string)
	var httpDescription string
	c := colly.NewCollector()
	c.OnHTML("div.col-16.col-xl-11.text.yellow.enable-copy.enable-external", func(e *colly.HTMLElement) {
		httpDescription = strings.Split(e.Text, ".")[0]
	})
	if err := c.Visit(fmt.Sprintf("https://http.dev/%v", httpCode)); err != nil {
		return mapDescription, err
	}
	mapDescription[httpCode] = httpDescription
	return mapDescription, nil
}
