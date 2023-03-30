package getDetails

import (
	"crawl-file/advancedCrawl"
	"crawl-file/connection"
	"crawl-file/model"
	"encoding/json"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"sync"
	"time"
)

const (
	redisQueue       = "update-domain-cubdomain"
	redisAddress     = "localhost:6379"
	urlBase          = "https://website.informer.com/"
	regexDescription = "<div id=\"description\">\\s*(.+?)\\s*</div>"
	regexKeywords    = "<div id=\"keywords\">\\n.*<b>Keywords:</b>\\n\\s*(.*?)\\s*</div>"
	regexTitle       = "<div id=\"title\">\\n\\s*(.*?)\\s*</div>"
	regexCreated     = "<td class=\"title\">Created:</td>\\n\\s*<td>(.*?)</td>"
	regexExpires     = "<td class=\"title\">Expires:</td>\\n\\s*<td>(.*?)</td>"
	regexOwner       = "td class=\"title\">Owner:</td>\\n\\s*<td>\\n\\s*<a.*>(.*?)</a>"
	layoutDate       = "2006-01-02"
)

var (
	Logger           *zap.Logger
	matchDescription = regexp.MustCompile(regexDescription)
	matchKeyword     = regexp.MustCompile(regexKeywords)
	matchTitle       = regexp.MustCompile(regexTitle)
	matchCreated     = regexp.MustCompile(regexCreated)
	matchExpires     = regexp.MustCompile(regexExpires)
	matchOwner       = regexp.MustCompile(regexOwner)
)

func GetRegexMatch(src string, regexName string) string {
	var re *regexp.Regexp
	if regexName == "description" {
		re = matchDescription
	}
	if regexName == "keywords" {
		re = matchKeyword
	}
	if regexName == "title" {
		re = matchTitle
	}
	if regexName == "created" {
		re = matchCreated
	}
	if regexName == "expires" {
		re = matchExpires
	}
	if regexName == "owner" {
		re = matchOwner
	}
	matches := re.FindAllStringSubmatch(src, -1)
	if matches != nil {
		return matches[0][1]
	}

	return "none"
}

func ConvertTime(dateStr string) time.Time {
	date, err := time.Parse(layoutDate, dateStr)
	if err != nil {
		var t time.Time //default time
		return t
	}
	return date
}

func GetDomainDetail(domainName string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	src, err := advancedCrawl.GetBody(*resp, err)
	if err != nil {
		return err
	}

	var domain = model.Domain{
		DomainUrl:   domainName, // string url lấy được
		Title:       GetRegexMatch(src, "title"),
		Description: GetRegexMatch(src, "title"),
		Keywords:    GetRegexMatch(src, "keywords"),
		Owner:       GetRegexMatch(src, "owner"),
		Expires:     ConvertTime(GetRegexMatch(src, "expires")),
		Created:     ConvertTime(GetRegexMatch(src, "created")),
	}

	err = connection.UpdateDataMongodb(domain)
	if err != nil {
		return err
	}

	return nil
}

func LoopInChan(chListDomains chan model.Domain) {
	for domain := range chListDomains {
		err := GetDomainDetail(domain.DomainUrl, urlBase+domain.DomainUrl)
		if err != nil {
			Logger.Error(err.Error())
			continue
		}
	}
}

func getListDomain(result []string, chListDomains chan<- model.Domain, limit int) {

	for i := 0; i < limit; i++ {
		var domain model.Domain

		err := json.Unmarshal([]byte(result[i]), &domain)
		if err != nil {
			Logger.Error(err.Error())
		}

		chListDomains <- domain
	}

	close(chListDomains)
}

func UploadDomains() error {

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	result, err := client.LRange(redisQueue, 0, -1).Result()
	if err != nil {
		return err
	}

	// Convert strings to Domains structs
	listDomains := make([]model.Domain, len(result))
	for i, v := range result {
		var domain model.Domain
		err = json.Unmarshal([]byte(v), &domain)
		if err != nil {
			Logger.Error(err.Error())
			continue
		}
		listDomains[i] = domain
	}

	limit := len(result)
	chListDomains := make(chan model.Domain)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		LoopInChan(chListDomains)
	}()

	go getListDomain(result, chListDomains, limit)
	wg.Wait()

	return nil
}
