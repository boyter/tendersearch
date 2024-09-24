package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/boyter/tendersearch/internal/common"
	"github.com/boyter/tendersearch/internal/database"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"log/slog"
	"strings"
	"time"
)

// RSS info
type RSS struct {
	Source      string `json:"source"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	UnixDate    int64  `json:"unixDate"`
	Categories  string `json:"categories"`
	Description string `json:"description"`
	Guid        string `json:"guid"`
	TimeAgo     string `json:"timeAgo"`
	Category    string `json:"category"`
}

// where a generic feed is identified put it here which will work in the short term but should not be considered
// a long term solution, where we prefer to perform a proper crawl and fetch of the source data
// so add new ones here where identified, but consider that a stop gap solution
var genericFeeds = []string{
	"https://www.tenders.gov.au/public_data/rss/rss.xml",
}

func (s *Service) processFeeds() {
	queries := database.New(s.db)

	for {
		slog.Info("starting processFeeds", common.UC("166e5ee0"))
		added := 0

		for _, feed := range genericFeeds {
			r := genericFetchFeed(feed)
			for _, res := range r {

				c, err := queries.TenderExistsByLink(context.Background(), res.Link)
				if err != nil {
					slog.Error("error", common.Err(err), "url", res.Link, common.UC("4284199e"))
					continue
				}

				// we already have this link so ignore it
				if c >= 1 {
					continue
				}

				jsonVersion := "v1"
				jsonText := "{}"
				errorText := ""

				// the following will actually go and fetch all the details for this source
				switch res.Source {
				case common.AusTenderCurrentATMList:
					t, err := fetchTendersGovAu(res.Link)
					if err != nil {
						errorText = err.Error()
						slog.Error("error", common.Err(err), "url", res.Link, "source", res.Source, common.UC("24d8c50e"))
					}
					m, _ := json.Marshal(t)
					jsonText = string(m)
				default:
					slog.Error("unknown source", "source", res.Source, common.UC("6e5d6946"))
				}

				tender := database.TenderCreateParams{
					Uuid:        uuid.NewString(),
					CreatedAt:   time.Now().Unix(),
					UpdatedAt:   time.Now().Unix(),
					Source:      res.Source,
					Link:        res.Link,
					Title:       res.Title,
					UnixDate:    res.UnixDate,
					Categories:  res.Categories,
					Description: res.Description,
					Guid:        res.Guid,
					JsonVersion: jsonVersion,
					Json:        jsonText,
					Attempt:     1,
					Error:       errorText,
				}

				tender = convertTenderParams(tender)

				_, err = queries.TenderCreate(context.Background(), tender)
				added++
				if err != nil {
					slog.Error("error", "err", err, common.UC("4227217a"))
				}
			}
		}

		slog.Info("sleeping processFeeds", "added", added, common.UC("0be69a43"))
		time.Sleep(15 * time.Minute)
	}
}

func convertTenderParams(tender database.TenderCreateParams) database.TenderCreateParams {
	switch tender.Source {
	case common.AusTenderCurrentATMList:
		var t TendersGovAuTenderV1
		err := json.Unmarshal([]byte(tender.Json), &t)
		if err != nil {
			slog.Error("error", common.Err(err), common.UC("b43fc398"))
		}

		_, tender.PublishAt = common.GuessDate(t.PublishDate, t.PublishDate)
		_, tender.ClosingAt = common.GuessDate(t.CloseDateTime, t.CloseDateTime)
	}

	return tender
}

// genericFetchFeed is a generic RSS fetcher that will pull down the feed and return
// what is hopefully good enough to get started before moving to something better
func genericFetchFeed(url string) []RSS {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fp := gofeed.NewParser()
	fp.UserAgent = common.UserAgent
	feed, err := fp.ParseURLWithContext(url, ctx)

	results := []RSS{}
	if err != nil {
		slog.Error("error", common.Err(err), common.UC("cb3635a1"))
		return results
	}

	for _, i := range feed.Items {
		unixTime := time.Now().Unix() // by default set to now

		if i.PublishedParsed != nil {
			unixTime = i.PublishedParsed.Unix()
		}
		if i.UpdatedParsed != nil {
			unixTime = i.UpdatedParsed.Unix()
		}

		results = append(results, RSS{
			Source:      feed.Title,
			Link:        i.Link,
			Title:       i.Title,
			Description: i.Description,
			Guid:        i.GUID,
			UnixDate:    unixTime,
			Categories:  strings.Join(i.Categories, " "),
		})
	}

	return results
}

// TendersGovAuTenderV1 represents a scraped version of a tender found on https://www.tenders.gov.au/
// such as https://www.tenders.gov.au/Atm/Show/654dd3b0-f4fe-4b04-a419-ecf0be09e586
type TendersGovAuTenderV1 struct {
	AtmId, Agency, Category, CloseDateTime, PublishDate, Location, AtmType                                   string
	AppReference, MultiAgencyAccess, MultiAgencyAccessType, PanelArrangement, MultiStage, MultiStageCriteria string
	Description, OtherInstructions                                                                           string
	ConditionsForParticipation, TimeframeForDelivery, EstimatedValue, AddressForLodgement                    string
}

// https://www.zenrows.com/blog/web-scraping-golang
// works on https://www.tenders.gov.au/public_data/rss/rss.xml
func fetchTendersGovAu(url string) (TendersGovAuTenderV1, error) {
	if !strings.Contains(url, "//www.tenders.gov.au/") {
		return TendersGovAuTenderV1{}, errors.New("url does not match expected format")
	}

	c := colly.NewCollector()
	// setting a valid User-Agent header so we can avoid detection
	c.UserAgent = common.UserAgent

	tender := TendersGovAuTenderV1{}

	c.OnHTML("div.listInner", func(e *colly.HTMLElement) {
		e.ForEach("div", func(i int, element *colly.HTMLElement) {
			// debug information printing
			//fmt.Println("SPAN>", element.ChildText("span"))
			//fmt.Println(" DIV>", element.ChildText("div"))

			var caseCovered bool
			m := strings.ToLower(element.ChildText("span"))

			// special case because this is mangled a bit in the output
			if strings.Contains(m, "close date & time") {
				tender.CloseDateTime = element.ChildText("div")
				// technically its 22 characters... BUT to be sure just trim
				if len(tender.CloseDateTime) > 24 {
					tender.CloseDateTime = strings.TrimSpace(tender.CloseDateTime[:24])
				}
				caseCovered = true
			}

			// another special case in the situation we don't match on the aud portion
			if strings.Contains(m, "estimated value") {
				tender.EstimatedValue = element.ChildText("div")
				caseCovered = true
			}

			switch m {
			case "atm id:":
				tender.AtmId = element.ChildText("div")
			case "agency:":
				tender.Agency = element.ChildText("div")
			case "category:":
				tender.Category = element.ChildText("div")
			case "close date & time:":
				tender.CloseDateTime = element.ChildText("div")
			case "publish date:":
				tender.PublishDate = element.ChildText("div")
			case "location:":
				tender.Location = element.ChildText("div")
			case "atm type:":
				tender.AtmType = element.ChildText("div")
			case "app reference:":
				tender.AppReference = element.ChildText("div")
			case "multi agency access":
				fallthrough
			case "multi agency access:":
				tender.MultiAgencyAccess = element.ChildText("div")
			case "multi agency access type":
				fallthrough
			case "multi agency access type:":
				tender.MultiAgencyAccessType = element.ChildText("div")
			case "panel arrangement:":
				tender.PanelArrangement = element.ChildText("div")
			case "multi-stage:":
				tender.MultiStage = element.ChildText("div")
			case "multi-stage criteria:":
				tender.MultiStageCriteria = element.ChildText("div")
			case "description:":
				tender.Description = element.ChildText("div")
			case "other instructions:":
				tender.OtherInstructions = element.ChildText("div")
			case "conditions for participation:":
				tender.ConditionsForParticipation = element.ChildText("div")
			case "timeframe for delivery:":
				tender.TimeframeForDelivery = element.ChildText("div")
			case "address for lodgement:":
				tender.AddressForLodgement = element.ChildText("div")

			case "": // no-op cases below till default
			case "addenda available":
			case "addenda available:":
			case "(act local time)":
			default:
				if !caseCovered {
					slog.Info("unknown fetchTendersGovAu case", "key", m, "url", url, common.UC("5cd180e5"))
				}
			}

		})
	})

	c.OnError(func(r *colly.Response, e error) {
		slog.Error("error", "err", e, "url", r.Request.URL.String(), common.UC("7c215b9c"))
	})

	err := c.Visit(url)
	if err != nil {
		return TendersGovAuTenderV1{}, err
	}

	return tender, nil
}
