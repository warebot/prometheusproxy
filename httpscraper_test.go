package prometheusproxy

import (
	helper "github.com/warebot/prometheusproxy/testhelpers"
	"net/url"
	"testing"
)

func Test_ScrapeOfflineEndpoint(t *testing.T) {
	scraper := NewHTTPScraper()
	url, _ := url.Parse("http://localhost")

	e := Endpoint{URL: url}
	msgs, errors, err := scraper.Scrape(e)
	if err != nil {
		panic(err)
	}

	msgsCount := 0
	go func() {
		for _ = range msgs {
			msgsCount++
		}
	}()

	errCount := 0
	for _ = range errors {
		errCount++
	}

	if errCount < 1 {
		t.Error(helper.AssertError("errCount > 1", errCount))
	}

}
