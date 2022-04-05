package domain

type Channel struct {
	ApplicationName ApplicationName
	Url             Url
	Rating          Rating
	NumberOfRatings RatingsAmount
}

func NewChannel(name ApplicationName, url Url, rating Rating, numberOfRating RatingsAmount) *Channel {
	return &Channel{ApplicationName: name, Url: url, Rating: rating, NumberOfRatings: numberOfRating}
}
