package getDetails

import (
	"crawl-file/advancedCrawl"
	"crawl-file/connection"
	"crawl-file/model"
	"encoding/json"
	"github.com/go-redis/redis"
	"log"
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
)

func GetRegexMatch(src string, regexName string) string {
	var re *regexp.Regexp
	if regexName == "description" {
		re = regexp.MustCompile(regexDescription)
	}
	if regexName == "keywords" {
		re = regexp.MustCompile(regexKeywords)
	}
	if regexName == "title" {
		re = regexp.MustCompile(regexTitle)
	}
	if regexName == "created" {
		re = regexp.MustCompile(regexCreated)
	}
	if regexName == "expires" {
		re = regexp.MustCompile(regexExpires)
	}
	if regexName == "owner" {
		re = regexp.MustCompile(regexOwner)
	}
	matches := re.FindAllStringSubmatch(src, -1)
	if matches != nil {
		return matches[0][1]
	}

	return "none"
}
func ConvertTime(dateStr string) time.Time {
	layout := "2006-01-02"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		var t time.Time //default time
		return t
	}
	return date
}
func GetDomainDetail(domainName string, url string) {
	resp, err := http.Get(url)
	src, err := advancedCrawl.GetBody(*resp, err)
	var domain = model.Domain{
		DomainUrl:   domainName, // string url lấy được
		Title:       GetRegexMatch(src, "title"),
		Description: GetRegexMatch(src, "title"),
		Keywords:    GetRegexMatch(src, "keywords"),
		Owner:       GetRegexMatch(src, "owner"),
		Expires:     ConvertTime(GetRegexMatch(src, "expires")),
		Created:     ConvertTime(GetRegexMatch(src, "created")),
	}
	err = connection.UpdateDataMongodb(advancedCrawl.Config, domain)
}
func LoopInChan(chListDomains chan model.Domain) {
	var wg sync.WaitGroup
	for domain := range chListDomains {
		GetDomainDetail(domain.DomainUrl, urlBase+domain.DomainUrl)
		wg.Add(1)
		wg.Done()
	}
	wg.Wait()

}
func UploadDomains() {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	result, err := client.LRange(redisQueue, 0, -1).Result()
	if err != nil {
		log.Fatal(err)
	}
	// Convert strings to User structs
	listDomains := make([]model.Domain, len(result))
	for i, v := range result {
		var domain model.Domain
		err = json.Unmarshal([]byte(v), &domain)
		if err != nil {
			log.Fatal(err)
		}
		listDomains[i] = domain
	}

	pageLimit := len(result)
	chListDomains := make(chan model.Domain, 100000)
	var wg sync.WaitGroup
	wg.Add(pageLimit)

	for i := 0; i < pageLimit; i++ {
		go func(taskNum int) {
			var domain model.Domain
			err = json.Unmarshal([]byte(result[taskNum]), &domain)
			if err != nil {
				log.Fatal(err)
			}
			defer wg.Done()
			chListDomains <- domain
		}(i)
	}
	wg.Wait()
	close(chListDomains)
	LoopInChan(chListDomains)
}
