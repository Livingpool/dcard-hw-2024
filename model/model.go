package model

import (
	"time"
)

type Gender string
type Platform string

const (
	Male    Gender   = "M"
	Female  Gender   = "F"
	Android Platform = "android"
	IOS     Platform = "ios"
	Web     Platform = "web"
)

// 產生廣告
type CreateAdRequest struct {
	Title      string `json:"title" validate:"required"`
	StartAt    string `json:"startAt" validate:"required"`
	EndAt      string `json:"endAt" validate:"required"`
	Conditions struct {
		AgeStart int        `json:"ageStart" validate:"min=1,max=100"`
		AgeEnd   int        `json:"ageEnd" validate:"min=1,max=100,ageRange"`
		Gender   []Gender   `json:"gender" validate:"dive,gender"`
		Country  []string   `json:"country" validate:"dive,countrycode"`
		Platform []Platform `json:"platform" validate:"dive,platform"`
	} `json:"conditions"`
}

// 列出符合可用和匹配目標條件的廣告
type SearchAdRequest struct {
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
	Conditions struct {
		Age      int      `json:"age" validate:"min=1,max=100"`
		Gender   Gender   `json:"gender" validate:"gender"`
		Country  string   `json:"country" validate:"countrycode"`
		Platform Platform `json:"platform" validate:"platform"`
	} `json:"conditions"`
}

// SearchAdResponse 回傳廣告的標題和結束時間
type SearchAdResponse struct {
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}
