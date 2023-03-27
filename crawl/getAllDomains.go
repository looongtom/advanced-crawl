package crawl

import (
	//"crawl-file/advancedCrawl"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

var (
	Logger *zap.Logger
)

const (
	patternUrl    = "<div class=\"col-md-4\">\\n<a href=\"(.+?)\">"
	patternDomain = "<div class=\"col-md-4\">\\n.+?>(.+?)</a>"
)

func getBody(resp http.Response, err error) (string, error) {
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			Logger.Error(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return string(body), err
	}
	return string(body), err //src cua url
}
func getListDomainADay(url string) {
	resp, err := http.Get(url)
	src, err := getBody(*resp, err)
	re := regexp.MustCompile(patternUrl)
	matches := re.FindAllStringSubmatch(src, -1)
	if matches != nil {
		for _, match := range matches {
			fmt.Println(match[1])
			//advancedCrawl.HandleListDomain(match[1])
		}
	}
}

func GetPageLimit(urlStart string) int {
	var pageLimit = 1
	response, err := http.Get(urlStart + strconv.Itoa(pageLimit))
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		s.Find("a.page-link").Each(func(i int, s *goquery.Selection) {
			pageText := s.Text()
			if pageText != "Next" {
				pageNum, err := strconv.Atoi(pageText)
				if pageNum >= pageLimit {
					pageLimit = pageNum
				}
				if err != nil {
					return
				}
			}
		})
	})
	return pageLimit
}
func LoopDayList(urlBase string) {
	for {
		if urlBase[len(urlBase)-1:] == "/" {
			break
		}
		urlBase = urlBase[:len(urlBase)-1]
	}
	pageLimit := GetPageLimit(urlBase)
	for i := 1; i <= pageLimit; i++ {
		pageNum := strconv.Itoa(i)
		fmt.Println("Page: " + pageNum)
		getListDomainADay(urlBase + pageNum)
		//time.Sleep(5 * time.Second)
	}
}
func HandleListDays(urlBase string) {
	for {
		if urlBase[len(urlBase)-1:] == "/" {
			break
		}
		urlBase = urlBase[:len(urlBase)-1]
	}
	pageLimit := GetPageLimit(urlBase)

	// create a wait group to ensure all goroutines have finished before exiting
	var wg sync.WaitGroup
	wg.Add(pageLimit)
	// launch 20 goroutines to perform the tasks
	for i := 0; i < pageLimit; i++ {
		go func(taskNum int) {

			defer wg.Done()
			fmt.Printf("Starting task %d\n", taskNum)
			for j := 0; j < 20; j++ {
				fmt.Printf("Task %d, sub-task %d\n", taskNum, j)
			}
			fmt.Printf("Finished task %d\n", taskNum)
		}(i)
	}

	// wait for all goroutines to finish
	wg.Wait()

	fmt.Println("All tasks completed")
}
