package reporting

import (
	"encoding/json"
	"log"
	"net/url"
	"tjweldon/spider/src/messaging"
)

func DomainsReport(backlog messaging.Backlog[string]) <-chan string {
	worker := func(incoming <-chan string, resultChan chan<- string) {
		defer close(resultChan)
		domains := map[string][]string{}
		for msg := range incoming {
			parsed, err := url.Parse(msg)
			if err != nil {
				continue
			}

			if paths, ok := domains[parsed.Host]; ok {
				domains[parsed.Host] = append(paths, parsed.Path)
			} else {
				domains[parsed.Host] = []string{parsed.Path}
			}
		}

		result, err := json.Marshal(&domains)
		if err != nil {
			log.Fatal(err)
		}
		resultChan <- string(result)
	}

	output := make(chan string)
	go worker(backlog.Channel(), output)

	return output
}
