package advancedCrawl

import (
	"crawl-file/connection"
	"crawl-file/model"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	Logger       *zap.Logger
	reListDay    = regexp.MustCompile(patternUrl)
	reListDomain = regexp.MustCompile(patternDomain)
	rePageLimit  = regexp.MustCompile(patternPagelimit)
)

const (
	patternUrl       = "<div class=\"col-md-4\">\\n( )*<a href=\"(.+?)\">"
	patternDomain    = "<div class=\"col-md-4\">\\n.+?>(.+?)</a>"
	patternPagelimit = "(\\d+)</a></li>\\n<li class=\"page-item\"><a class=\"page-link\" .+xt</a></li>"
)

func GetMatchesByRegex(s string, chList chan<- string) {
	matches := reListDay.FindAllStringSubmatch(s, -1)

	var wg sync.WaitGroup

	if matches != nil {

		for i := 0; i < len(matches); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				if len(matches[index]) > 1 {
					chList <- matches[index][2] // --> panic
				}

			}(i)
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

func GetPageLimit(urlStart string) (int, error) {
	resp, err := http.Get(urlStart)
	if err != nil {
		return 1, err // nếu xảy ra lỗi không lấy được pageLimit thì cho = 1
	}

	src, err := GetBody(*resp, err)
	if err != nil {
		return 1, err // nếu xảy ra lỗi không lấy được pageLimit thì cho = 1
	}
	matches := rePageLimit.FindStringSubmatch(src)
	var pageLimit int
	if len(matches) > 1 {
		match := matches[1]
		pageLimit, err = strconv.Atoi(match)
		if err != nil {
			Logger.Error(err.Error())
			return 1, err // nếu xảy ra lỗi không lấy được pageLimit thì cho = 1
		}
	}
	return pageLimit, nil
}

func GetListDomainInADay(url string, chListDay chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		Logger.Error(err.Error())
	}
	src, err := GetBody(*resp, err)
	if err != nil {
		Logger.Error(err.Error())
	}
	GetMatchesByRegex(src, chListDay)
}

func getListDay(urlBase string, pageLimit int, chListDay chan<- string) {
	for page := 1; page <= pageLimit; page++ {
		url := fmt.Sprintf("%s%d", urlBase, page)
		GetListDomainInADay(url, chListDay)
	}
	close(chListDay)
}

func GetMatchedDomains(s string) {
	matches := reListDomain.FindAllStringSubmatch(s, -1)
	var wg sync.WaitGroup
	var doc []interface{}

	if matches != nil {
		for i := 0; i < len(matches); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				if len(matches[index]) > 0 {

					var defaultTime time.Time // default time
					var D = model.Domain{
						DomainUrl:   matches[index][1], // string url lấy được
						Title:       "",
						Description: "",
						Keywords:    "",
						Owner:       "",
						Expires:     defaultTime,
						Created:     defaultTime,
					}

					doc = append(doc, D)

				}
			}(i)
		}
		wg.Wait()
	}

	fmt.Println("Saving in mongo ", len(doc), "domains ")

	models := make([]mongo.WriteModel, len(doc))
	for i, domain := range doc {
		models[i] = mongo.NewInsertOneModel().SetDocument(domain)
	}

	err := connection.SaveFileToMongoDb(models)
	if err != nil {
		Logger.Error(err.Error())
	}
}

func GetListDomainInAPage(url string) {
	resp, err := http.Get(url)
	if err != nil {
		Logger.Error(err.Error())
	}

	src, err := GetBody(*resp, err)
	if err != nil {
		Logger.Error(err.Error())
	}

	GetMatchedDomains(src)
}

func getListDomainThroughPages(urlDay string, pageLimit int) {
	if len(urlDay) > 0 {
		urlDay = urlDay[:len(urlDay)-1]
	} else {
		Logger.Error("therse is nothing in " + urlDay)
	}
	for page := 1; page <= pageLimit; page++ {
		url := fmt.Sprintf("%s%d", urlDay, page)
		GetListDomainInAPage(url)
	}
}

func getListDomain(chListDay chan string) {
	for day := range chListDay {
		pageLimit, err := GetPageLimit(day)
		if err != nil {
			Logger.Error(err.Error())
		}
		getListDomainThroughPages(day, pageLimit)
	}
}

func HandleListDomain(urlBase string) error {
	pageLimit, err := GetPageLimit(urlBase)
	if err != nil {
		return err
	}
	chListDay := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		getListDomain(chListDay)
	}()

	go getListDay(urlBase, pageLimit, chListDay)
	wg.Wait()
	return nil
}
