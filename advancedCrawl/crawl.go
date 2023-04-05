package advancedCrawl

import (
	"crawl-file/model"
	"crawl-file/service"
	"errors"
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
				if len(matches[index]) < 3 {
					return
				}

				chList <- matches[index][2] // --> panic
			}(i)
		}

		wg.Wait()
	}
}

func GetBody(resp http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	return string(body), err //src cua url
}

func GetPageLimit(urlStart string) (int, error) {
	resp, err := http.Get(urlStart)
	if err != nil {
		return 1, err // nếu xảy ra lỗi không lấy được pageLimit thì cho = 1
	}
	defer resp.Body.Close()

	src, err := GetBody(*resp)
	if err != nil {
		return 1, err // nếu xảy ra lỗi không lấy được pageLimit thì cho = 1
	}

	matches := rePageLimit.FindStringSubmatch(src)
	var pageLimit int
	if len(matches) < 2 {
		return 0, fmt.Errorf("errror")
	}

	match := matches[1]
	pageLimit, err = strconv.Atoi(match)
	if err != nil {
		return 1, err
	}

	return pageLimit, nil
}

func GetListDomainInADay(url string, chListDay chan<- string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	src, err := GetBody(*resp)
	if err != nil {
		return err
	}

	GetMatchesByRegex(src, chListDay)
	return nil
}

func getListDay(urlBase string, pageLimit int, chListDay chan<- string) {
	for page := 1; page <= pageLimit; page++ {
		url := fmt.Sprintf("%s%d", urlBase, page)
		err := GetListDomainInADay(url, chListDay)
		if err != nil {
			Logger.Error(err.Error())
			continue
		}
	}
	close(chListDay)
}

func GetMatchedDomains(s string) error {
	matches := reListDomain.FindAllStringSubmatch(s, -1)
	var wg sync.WaitGroup
	var doc []interface{}

	if matches != nil {
		for i := 0; i < len(matches); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				if len(matches[index]) < 2 {
					return
				}

				var defaultTime time.Time // default time
				var D = model.Domain{
					DomainUrl:   matches[index][1], // string url lấy được
					Title:       "",
					Description: "",
					Keywords:    "",
					Owner:       "",
					Expires:     defaultTime,
					Created:     defaultTime,
					Status:      model.StatusDisable,
				}

				doc = append(doc, D)
			}(i)
		}
		wg.Wait()
	}

	models := make([]mongo.WriteModel, len(doc))
	for i, domain := range doc {
		models[i] = mongo.NewInsertOneModel().SetDocument(domain)
	}

	return service.SaveFileToMongoDb(models)
}

func GetListDomainInAPage(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	src, err := GetBody(*resp)
	if err != nil {
		return err
	}

	err = GetMatchedDomains(src)
	if err != nil {
		return err
	}
	return nil
}

func getListDomainThroughPages(urlDay string, pageLimit int) error {
	if len(urlDay) < 0 {
		return errors.New("invalid url" + urlDay)
	}
	urlDay = urlDay[:len(urlDay)-1]

	for page := 1; page <= pageLimit; page++ {
		url := fmt.Sprintf("%s%d", urlDay, page)
		err := GetListDomainInAPage(url)
		if err != nil {
			return err
		}
	}
	return nil
}

func getListDomain(chListDay chan string, wg *sync.WaitGroup) {
	for day := range chListDay {
		time.Sleep(time.Second * 5)
		pageLimit, err := GetPageLimit(day)

		if err != nil {
			fmt.Println(err)
			continue
		}

		err = getListDomainThroughPages(day, pageLimit)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	wg.Done()
}

func HandleListDomain(urlBase string) error {
	pageLimit, err := GetPageLimit(urlBase)
	if err != nil {
		return err
	}
	chListDay := make(chan string)
	noOfWorkers := 10
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go getListDomain(chListDay, &wg)
	}
	go getListDay(urlBase, pageLimit, chListDay)
	wg.Wait()
	return nil
}
