package lib

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type S3Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	HostId    string   `xml:"HostId"`
	RequestId string   `xml:"RequestId"`
}

func S3ErrorIsRequestHasExpired(err error) bool {
	var s3Error S3Error
	return errors.As(err, &s3Error) && s3Error.Message == "Request has expired"
}

func S3ErrorIsRequestTimeout(err error) bool {
	var s3Error S3Error
	return errors.As(err, &s3Error) && s3Error.Code == "RequestTimeout"
}

func (s S3Error) Error() string {
	return fmt.Sprintf("%v - %v", s.Code, s.Message)
}

func (s S3Error) Empty() bool {
	return s.Message == "" && s.Code == ""
}
