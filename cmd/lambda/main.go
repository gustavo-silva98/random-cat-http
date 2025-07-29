package main

import (
	"encoding/json"
	"fmt"
	"log"
	"random-http-cat/internal/cat"
	"random-http-cat/internal/mdn"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	//"random-http-cat/pkg/randomizer"
	"random-http-cat/internal/dynamo"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	timeIni := time.Now()
	mapCode, err := populateMapScrap()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Extração feita. Número de elementos: %v\n", len(mapCode))
	}
	fmt.Println("Printando AWS")
	if _, err := dynamo.CreateHttpTable(); err != nil {
		log.Println(err)
	}
	dynamo.ListTable()
	var putRequestSlice []*dynamodb.WriteRequest
	count := 0
	session := dynamo.GetSession()
	for key, value := range mapCode {
		tempMap := map[int]string{
			key: value,
		}
		kbs, err := evaluateKBs(tempMap)
		switch {
		case err != nil:
			log.Println(err)
		case kbs > 400:
			log.Println("Tamanho grande demais para inserir no DB:", kbs)
		default:
			dynamo.AddPutRequestSlice(&putRequestSlice, key, value)
			count++
		}
		if count%25 == 0 {
			fmt.Println("Número de itens a serem inseridos", len(putRequestSlice))
			dynamo.BatchWriteItem(session, &putRequestSlice, "httpDescription")
			putRequestSlice = nil
		}
	}
	if len(putRequestSlice) > 0 {
		fmt.Println("Número de itens a serem inseridos", len(putRequestSlice))
		dynamo.BatchWriteItem(session, &putRequestSlice, "httpDescription")
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
		switch {
		case err != nil:
			log.Fatalf("[CRITICAL] - Erro ao extrair %v", err)
		case desc[job] == "":
			log.Fatalf("[CRITICAL]- Empty: Extração do code %v veio vazia.\n", job)
		default:
			results <- desc
		}
	}
}

func evaluateKBs(data map[int]string) (float64, error) {
	avMap, err := attributevalue.MarshalMap(data)
	if err != nil {
		return 0, err
	}

	jsonBytes, err := json.Marshal(avMap)
	if err != nil {
		return 0, err
	}

	return float64(len(jsonBytes)), nil
}
