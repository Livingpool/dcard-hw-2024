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
	Title      string `json:"title" bson:"title" validate:"required"`
	StartAt    string `json:"startAt" bson:"startAt" validate:"required"`
	EndAt      string `json:"endAt" bson:"endAt" validate:"required"`
	Conditions `json:"conditions" bson:"conditions"`
}

type Conditions struct {
	AgeStart int        `json:"ageStart" bson:"ageStart" validate:"min=0,max=100"`
	AgeEnd   int        `json:"ageEnd" bson:"ageEnd" validate:"min=0,max=100,ageRange"`
	Gender   []Gender   `json:"gender" bson:"gender" validate:"dive,gender"`
	Country  []string   `json:"country" bson:"country" validate:"dive,countrycode"`
	Platform []Platform `json:"platform" bson:"platform" validate:"dive,platform"`
}

type CreateAdInMongoDB struct {
	Title      string    `json:"title" bson:"title" validate:"required"`
	StartAt    time.Time `json:"startAt" bson:"startAt" validate:"required"`
	EndAt      time.Time `json:"endAt" bson:"endAt" validate:"required"`
	Conditions `json:"conditions" bson:"conditions"`
}

// 列出符合可用和匹配目標條件的廣告
type SearchAdRequest struct {
	Offset   int      `json:"offset" form:"offset"`
	Limit    int      `json:"limit" form:"limit" validate:"max=100"`
	Age      int      `json:"age" form:"age" validate:"min=0,max=100"`
	Gender   Gender   `json:"gender" form:"gender" validate:"gender"`
	Country  string   `json:"country" form:"country" validate:"countrycode"`
	Platform Platform `json:"platform" form:"platform" validate:"platform"`
}

// SearchAdResponse 回傳廣告的標題和結束時間
type SearchAdResponse struct {
	Title string    `json:"title" bson:"title"`
	EndAt time.Time `json:"endAt" bson:"endAt"`
}
