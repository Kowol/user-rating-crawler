package domain

import (
	"errors"
	"fmt"
	"net/url"
)

type Url string
type ApplicationName string
type Rating float32
type RatingsAmount uint32 // Not sure how many rating it could have but at least we know it will be a positive number

func NewURL(value string) (*Url, error) {
	if value == "" {
		return nil, errors.New("url could not be empty")
	}

	_, err := url.ParseRequestURI(value)
	if err != nil {
		return nil, fmt.Errorf("invalid url specified %w", err)
	}

	crawlUrl := Url(value)
	return &crawlUrl, nil
}

func NewApplicationName(value string) (*ApplicationName, error) {
	if value == "" {
		return nil, errors.New("application name could not be empty")
	}

	appName := ApplicationName(value)
	return &appName, nil
}

func NewRating(value float32) (*Rating, error) {
	if value < 0 {
		return nil, errors.New("rating has to be positive number")
	}

	rating := Rating(value)

	return &rating, nil
}

func NewRatingsAmount(value uint32) (*RatingsAmount, error) {
	ratingAmount := RatingsAmount(value)

	return &ratingAmount, nil
}
