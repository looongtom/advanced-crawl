package advancedCrawl

import (
	"crawl-file/connection"
	"crawl-file/dataConfig"
	"crawl-file/model"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	Logger          *zap.Logger
	checkListDay    = true
	checkListDomain = false
	Config          *dataConfig.Config
	err             error
)

const (
	patternUrl    = "<div class=\"col-md-4\">\\n( )*<a href=\"(.+?)\">"
	patternDomain = "<div class=\"col-md-4\">\\n.+?>(.+?)</a>"
)

func init() {
	Config, err = connection.ReadEnv()
	if err != nil {
		Logger.Error(err.Error())
	}
}
func GetMatchedDomainsByRegex(s string, url string) {
	var re *regexp.Regexp
	if checkListDay {
		re = regexp.MustCompile(patternUrl)
	}
	if checkListDomain {
		re = regexp.MustCompile(patternDomain)
	}
	matches := re.FindAllStringSubmatch(s, -1)
	var doc []interface{}
	var wg sync.WaitGroup
	if matches != nil {
		for _, match := range matches {
			var defaultTime time.Time // default time
			var D = model.Domain{
				DomainUrl:   match[1], // string url lấy được
				Title:       "",
				Description: "",
				Keywords:    "",
				Owner:       "",
				Expires:     defaultTime,
				Created:     defaultTime,
			}
			temp := bson.M{
				"domain":      D.GetDomainUrl(),
				"title":       D.GetTitle(),
				"description": D.GetDescription(),
				"keywords":    D.GetKeywords(),
				"owner":       D.GetOwner(),
				"expires":     D.GetExpires(),
				"created":     D.GetCreated(),
			}
			doc = append(doc, temp)
			if err != nil {
				Logger.Error(err.Error())
			}

			wg.Add(1)
			wg.Done()

		}
		wg.Wait()
	}
	fmt.Println("Saving in mongo ", len(doc), "domains")
	models := make([]mongo.WriteModel, len(doc))
	for i, domain := range doc {
		models[i] = mongo.NewInsertOneModel().SetDocument(domain)
	}
	err = connection.SaveFileToMongoDb(Config, models)
}
func GetMatchesByRegex(s string, chList chan string) {
	var re *regexp.Regexp
	if checkListDay {
		re = regexp.MustCompile(patternUrl)
	}
	if checkListDomain {
		re = regexp.MustCompile(patternDomain)
	}
	matches := re.FindAllStringSubmatch(s, -1)
	var wg sync.WaitGroup

	if matches != nil {
		for _, match := range matches {
			wg.Add(1)
			wg.Done()
			if checkListDay {
				chList <- match[2]
			}
			if checkListDomain {
				chList <- match[1]
			}
		}
		wg.Wait()

	}
}
func GetBody(resp http.Response, err error) (string, error) {
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
func GetPageLimit(urlStart string) int {
	var pageLimit = 1
	response, err := http.Get(urlStart + strconv.Itoa(pageLimit))
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(response.Body)
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
func HandleALotOfDomains(chListDay chan string) {
	numWorkers := 6
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	// Launch workers
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for day := range chListDay {
				pageLimit := GetPageLimit(day)
				for {
					if day[len(day)-1:] == "/" {
						break
					}
					day = day[:len(day)-1]
					var subWg sync.WaitGroup
					for i := 1; i <= pageLimit; i++ {
						pageNum := strconv.Itoa(i)
						GetAllDomainInADayList(day + pageNum)
						subWg.Add(1)
						subWg.Done()
					}
					subWg.Wait()
				}
			}
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()

}

func GetListDomainInADay(url string, chListDay chan string) {
	resp, err := http.Get(url)
	src, err := GetBody(*resp, err)
	GetMatchesByRegex(src, chListDay)
}
func GetAllDomainInADayList(url string) {
	resp, err := http.Get(url)
	src, err := GetBody(*resp, err)
	GetMatchedDomainsByRegex(src, url)
}
func HandleListDomain(urlBase string) {
	//chạy lần đầu checkListDay =true => Đánh dấu: lấy danh sách các ngày
	pageLimit := GetPageLimit(urlBase)
	for {
		if urlBase[len(urlBase)-1:] == "/" {
			break
		}
		urlBase = urlBase[:len(urlBase)-1]
	}
	chListDay := make(chan string, 100000)
	var wg sync.WaitGroup
	wg.Add(pageLimit)

	for i := 1; i <= pageLimit; i++ {
		go func(taskNum int) {
			pageNum := strconv.Itoa(taskNum)
			defer wg.Done()
			GetListDomainInADay(urlBase+pageNum, chListDay)
		}(i)
	}

	wg.Wait()
	close(chListDay)
	checkListDay = false
	checkListDomain = true //Đánh dấu: Lấy danh sách các domains trong 1 ngày
	HandleALotOfDomains(chListDay)
}

//func main() {
//	checkListDay = true
//	HandleListDomain("https://www.cubdomain.com/domains-registered-dates/1")
//}
