package main

import (
	"fmt"
	"random-http-cat/internal/cat"
	"random-http-cat/internal/mdn"
	"sync"
	"time"
	//"random-http-cat/pkg/randomizer"
)

func main() {
	timeIni := time.Now()
	mapCode, err := populateMapScrap()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Extração feita. Número de elementos: %v\n", len(mapCode))
	}
	fmt.Println(time.Since(timeIni))
}

func populateMapScrap() (map[int]string, error) {
	catCodes := cat.GetCatCodes()

	var wg sync.WaitGroup
	jobs := make(chan int, len(catCodes))
	results := make(chan map[int]string, len(catCodes))

	for w := 1; w <= len(catCodes); w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	for _, code := range catCodes {
		jobs <- code
	}
	close(jobs)
	wg.Wait()
	close(results)
	finalMap := make(map[int]string)
	for result := 1; result <= len(catCodes); result++ {
		mapWorker := <-results
		for key, value := range mapWorker {
			finalMap[key] = value
		}
	}
	if len(catCodes) == len(finalMap) {
		return finalMap, nil
	} else {
		return nil, fmt.Errorf("falha na extração. Itens esperados: %v - Itens extraídos: %v", len(catCodes), len(finalMap))
	}
}

func worker(_ int, jobs <-chan int, results chan<- map[int]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		desc, err := mdn.GetHttp(job)
		if err != nil {
		} else {
			results <- desc
		}
	}
}
