package downloadurl

import (
	"net/url"
	"strconv"
	"time"
)

type URL struct {
	*url.URL
	time.Time
	Type URLType
}

func New(urlStr string) (d *URL, err error) {
	d = &URL{}
	err = d.parseUrl(urlStr)
	if err != nil {
		return
	}

	err = d.parseTime()

	return
}

func (d *URL) parseUrl(urlStr string) (err error) {
	d.URL, err = url.Parse(urlStr)
	if err != nil {
		return
	}
	return
}

type expire struct {
	date       string
	expire     string
	expireDate string
	urlType    URLType
}

type URLType string

const (
	AmazonS3 URLType = "AmazonS3"
	Google   URLType = "Google"
	Azure    URLType = "Azure"
	Files    URLType = "Files"
)

var (
	amazonS3       = expire{"X-Amz-Date", "X-Amz-Expires", "", AmazonS3}
	filesDate      = expire{"X-Files-Date", "X-Files-Expires", "", Files}
	googleDate     = expire{"X-Goog-Date", "X-Goog-Expires", "", Google}
	azureBlob      = expire{"sp", "", "se", Azure}
	timeDateFormat = "20060102T150405Z"
)

var Dates = []expire{amazonS3, filesDate, googleDate, azureBlob}

func (d *URL) parseTime() (err error) {
	for _, parser := range Dates {
		date := d.URL.Query().Get(parser.date)
		if parser.expireDate != "" {
			query := (&url.URL{RawQuery: date}).Query()
			if len(query[parser.expireDate]) == 1 {
				d.Time, err = time.Parse(timeDateFormat, query[parser.expireDate][0])
				d.Type = parser.urlType
			} else {
				continue
			}
		} else {
			d.Time, err = time.Parse(timeDateFormat, date)
		}

		if parser.expire != "" {
			duration, err := strconv.Atoi(d.URL.Query().Get(parser.expire))
			if err == nil {
				t := time.Second * time.Duration(duration)
				d.Time = d.Time.Add(t)
			}
		}
		if err == nil {
			break
		}
	}
	return err
}

func (d *URL) Valid(within time.Duration) (remaining time.Duration, valid bool) {
	remaining = time.Until(d.Time)
	valid = remaining >= within
	return
}
