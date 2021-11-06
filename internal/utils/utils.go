package utils

import (
	"fmt"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/tidwall/gjson"
)

func GetUploadDate(picture string) int64 {
	value := gjson.Get(picture, "uploadDate")
	if value.Exists() {
		return value.Int() / 1000
	}
	return gjson.Get(picture, "upload_date").Int() / 1000
}

func Filename(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	s, err := strftime.Format("%Y-%m-%d_%H-%M-%S", tm)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s_UTC.jpg", s)
}
