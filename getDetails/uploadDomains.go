package getDetails

import (
	"crawl-file/advancedCrawl"
	"crawl-file/model"
	"crawl-file/service"
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
	redisClient      = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	client = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
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

	src, err := advancedCrawl.GetBody(*resp)
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

	err = service.UpdateDataMongodb(domain)
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

func UploadDomains() error {
	result, err := client.LRange(redisQueue, 0, -1).Result()
	if err != nil {
		return err
	}
	listDomains := make(chan model.Domain)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go LoopInChan(listDomains)
	}()

	go func() {
		for v := range result {
			var domain model.Domain
			err = json.Unmarshal([]byte(result[v]), &domain)
			if err != nil {
				Logger.Error(err.Error())
				continue
			}
			if domain.Status == model.StatusEnable {
				listDomains <- domain
			}

		}
		close(listDomains)
	}()
	wg.Wait()

	return nil
}
