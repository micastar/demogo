package bin

import (
	"github.com/go-playground/validator/v10"
	"github.com/mmcdole/gofeed"
)

type RSSReader struct {
	parser *gofeed.Parser
	url    string
	Feed   *gofeed.Feed
}

func NewRSSReader(baseURL string) *RSSReader {
	return &RSSReader{
		parser: gofeed.NewParser(),
		url:    baseURL,
	}
}

func (r *RSSReader) FetchFeed() error {
	err := checkURL(r.url)
	if err != nil {
		return &ErrMessage{Message: err.Error()}
	}
	feed, err := r.parser.ParseURL(r.url)
	if err != nil {
		return &ErrMessage{Message: err.Error()}
	}
	r.Feed = feed
	return nil

}

var validate *validator.Validate

func checkURL(url string) error {
	newURL := url
	validate = validator.New(validator.WithPrivateFieldValidation())
	return validate.Var(newURL, "url")
}

type ErrMessage struct {
	Message string
}

func (e *ErrMessage) Error() string {
	return e.Message
}
