package cat

import (
	"log"
	"strconv"

	"github.com/gocolly/colly/v2"
)

func GetCatCodes() []int {
	var catCodes []int
	c := colly.NewCollector()
	c.OnHTML(`div[class="text-[2rem] tracking-[2px] font-semibold uppercase"]`, func(e *colly.HTMLElement) {
		if code, err := strconv.Atoi(e.Text); err != nil {
			log.Fatalf("Falha ao converter para Int o code %v", err)
		} else {
			catCodes = append(catCodes, code)
		}

	})
	if err := c.Visit("https://http.cat/"); err != nil {
		log.Fatalf("Erro no Visit url: %v", err)
	}

	return catCodes
}
