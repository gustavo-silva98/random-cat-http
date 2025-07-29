package mdn

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
	"github.com/gocolly/colly/v2"
)

func GetHttp(httpCode int) (map[int]string, error) {

	mapDescription := make(map[int]string)
	var httpDescription string
	c := colly.NewCollector()
	c.OnHTML("div.col-16.col-xl-11.text.yellow.enable-copy.enable-external", func(e *colly.HTMLElement) {
		httpDescription = strings.Split(e.Text, ".")[0]
		if httpDescription == "" {
			fmt.Println("Pegou do DOM")
			e.DOM.Find("p").Each(func(_ int, s *goquery.Selection) {
				httpDescription = s.Text()
			})
		}
	})
	if err := c.Visit(fmt.Sprintf("https://http.dev/%v", httpCode)); err != nil {
		return mapDescription, err
	}
	mapDescription[httpCode] = httpDescription
	return mapDescription, nil
}

func GetHTML(httpCode int) {
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		//var reader io.Reader = bytes.NewReader(r.Body)

		fmt.Println("headers=", r.Headers.Get("Content-Encoding"))
		brReader := brotli.NewReader(bytes.NewReader(r.Body))
		body, err := io.ReadAll(brReader)
		if err != nil {
			log.Println("Erro ao ler conteudo Brotli:", err)
			return
		}
		fmt.Println(string(body))
	})
	if err := c.Visit(fmt.Sprintf("https://http.dev/%v", httpCode)); err != nil {
		log.Println(httpCode, "-", err)
	}

}
