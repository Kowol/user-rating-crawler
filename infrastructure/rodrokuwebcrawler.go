package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"go-web-crawler-service/domain"
	"log"
	"regexp"
	"strconv"
	"time"
)

const (
	crawlerTTLSeconds = 20
)

type rodRokuWebCrawler struct {
	browser *rod.Browser
}

func NewRodRokuWebCrawler(browser *rod.Browser) *rodRokuWebCrawler {
	return &rodRokuWebCrawler{
		browser: browser,
	}
}

func (c *rodRokuWebCrawler) CrawlChannel(ctx context.Context, url domain.Url) (*domain.Channel, error) {
	ctx, cancel := context.WithTimeout(ctx, crawlerTTLSeconds*time.Second)
	defer cancel()

	browser := c.browser.Context(ctx)

	var page *rod.Page
	err := rod.Try(
		func() {
			page = browser.MustPage(string(url))
		},
	)
	checkedErr := checkErr(err)
	if checkedErr != nil {
		return nil, checkedErr
	}
	defer func() {
		_ = rod.Try(
			func() {
				page.MustClose()
			},
		)
	}()

	log.Printf("Successfully opened page with url: %s\n", url)

	page = page.Context(ctx)

	var hero *rod.Element
	err = rod.Try(
		func() {
			hero = page.MustElement(".Roku-Page-Details-Hero")
		},
	)
	checkedErr = checkErr(err)
	if checkedErr != nil {
		return nil, checkedErr
	}

	appName, err := getAppName(hero)
	if err != nil {
		return nil, fmt.Errorf("failed to get application name, %w", err)
	}

	avgRating, err := getAvgRating(hero)
	if err != nil {
		return nil, fmt.Errorf("failed to get average rating, %w", err)
	}

	ratingsAmount, err := getRatingsAmount(hero)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings amount, %w", err)
	}

	channel, err := createChannel(url, appName, avgRating, ratingsAmount)
	if err != nil {
		return nil, fmt.Errorf("unable to create channel entity from scrapped data, error: %w", err)
	}

	return channel, nil
}

func checkErr(err error) error {
	var evalErr *rod.ErrEval
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timeout %w", err)
	} else if errors.As(err, &evalErr) {
		return fmt.Errorf("evaluation error, line: %d, error: %w", evalErr.LineNumber, evalErr)
	} else if err != nil {
		return fmt.Errorf("unable to scrap website, error: %w", err)
	}

	return nil
}

func getAppName(hero *rod.Element) (string, error) {
	var appNameElement *rod.Element
	var appName string

	err := rod.Try(
		func() {
			appNameElement = hero.MustElement("h1")
		},
	)
	checkedErr := checkErr(err)
	if checkedErr != nil {
		return "", checkedErr
	}

	err = rod.Try(
		func() {
			appName = appNameElement.MustText()
		},
	)
	checkedErr = checkErr(err)
	if checkedErr != nil {
		return "", checkedErr
	}

	return appName, nil
}

func getAvgRating(hero *rod.Element) (float32, error) {
	var avgRatingElement *rod.Element
	var avgRatingVal string

	err := rod.Try(
		func() {
			avgRatingElement = hero.MustElement(".average-rating")
		},
	)
	checkedErr := checkErr(err)
	if checkedErr != nil {
		return 0, checkedErr
	}

	err = rod.Try(
		func() {
			avgRatingVal = avgRatingElement.MustText()
		},
	)
	checkedErr = checkErr(err)
	if checkedErr != nil {
		return 0, checkedErr
	}

	avgRating, err := strconv.ParseFloat(avgRatingVal, 32)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to create float value from average rating, value: %s, error: %w",
			avgRatingVal,
			err,
		)
	}

	return float32(avgRating), nil
}

func getRatingsAmount(hero *rod.Element) (uint32, error) {
	var ratingsAmountCntElement *rod.Element
	var ratingsAmountCntVal string

	err := rod.Try(
		func() {
			ratingsAmountCntElement = hero.MustElement("[itemprop=\"starRating\"]")
		},
	)
	checkedErr := checkErr(err)
	if checkedErr != nil {
		return 0, checkedErr
	}

	err = rod.Try(
		func() {
			ratingsAmountCntVal = ratingsAmountCntElement.MustText()
		},
	)
	checkedErr = checkErr(err)
	if checkedErr != nil {
		return 0, checkedErr
	}

	var extractedRatingsAmount = "0"

	r := regexp.MustCompile(`(\d+) ratings`)
	extractedRating := r.FindStringSubmatch(ratingsAmountCntVal)
	if len(extractedRating) == 2 {
		extractedRatingsAmount = extractedRating[1]
	}

	ratingsAmount, err := strconv.Atoi(extractedRatingsAmount)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to create integer value from ratings amount, value: %d, error: %w",
			ratingsAmount,
			err,
		)
	}

	return uint32(ratingsAmount), nil
}

func createChannel(url domain.Url, nameVal string, ratingVal float32, ratingsAmountVal uint32) (
	*domain.Channel,
	error,
) {
	appName, err := domain.NewApplicationName(nameVal)
	if err != nil {
		return nil, fmt.Errorf("application name is invalid: %s, error: %w", nameVal, err)
	}

	rating, err := domain.NewRating(ratingVal)
	if err != nil {
		return nil, fmt.Errorf("rating has invalid value: %f, error: %w", ratingVal, err)
	}

	ratingsAmount, err := domain.NewRatingsAmount(ratingsAmountVal)
	if err != nil {
		return nil, fmt.Errorf("ratings amount has invalid value: %d, error: %w", ratingsAmountVal, err)
	}

	return domain.NewChannel(*appName, url, *rating, *ratingsAmount), nil
}
