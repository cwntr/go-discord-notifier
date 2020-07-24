package fourchan

import (
	"encoding/json"
	"net/http"
)

const (
	BIZCatalogURL    = "https://a.4cdn.org/biz/catalog.json"
	BIZThreadLinkURL = "https://boards.4channel.org/biz/thread"
)

type CatalogResponse []CatalogPageItems

type CatalogPageItems struct {
	Page    int         `json:"page"`
	Threads []APIThread `json:"threads"`
}

type Client struct {
	c       http.Client
	BaseURL string
}

func NewClient() *Client {
	return &Client{BaseURL: BIZCatalogURL}
}

func (c *Client) GetCatalog() (*[]CatalogPageItems, error) {
	res, err := c.c.Get(c.BaseURL)
	if err != nil {
		return nil, err
	}
	var resData []CatalogPageItems
	err = getJson(res, &resData)
	if err != nil {
		return nil, err
	}
	return &resData, nil
}

func getJson(response *http.Response, target interface{}) error {
	defer response.Body.Close()
	return json.NewDecoder(response.Body).Decode(target)
}

type APIThread struct {
	No            int         `json:"no"`
	Sticky        int         `json:"sticky,omitempty"`
	Closed        int         `json:"closed,omitempty"`
	Now           string      `json:"now"`
	Name          string      `json:"name"`
	Sub           string      `json:"sub,omitempty"`
	Com           string      `json:"com,omitempty"`
	Filename      string      `json:"filename"`
	Ext           string      `json:"ext"`
	W             int         `json:"w"`
	H             int         `json:"h"`
	TnW           int         `json:"tn_w"`
	TnH           int         `json:"tn_h"`
	Tim           int64       `json:"tim"`
	Time          int         `json:"time"`
	Md5           string      `json:"md5"`
	Fsize         int         `json:"fsize"`
	Resto         int         `json:"resto"`
	ID            string      `json:"id"`
	Capcode       string      `json:"capcode,omitempty"`
	SemanticURL   string      `json:"semantic_url"`
	Replies       int         `json:"replies"`
	Images        int         `json:"images"`
	LastModified  int         `json:"last_modified"`
	Bumplimit     int         `json:"bumplimit,omitempty"`
	Imagelimit    int         `json:"imagelimit,omitempty"`
	OmittedPosts  int         `json:"omitted_posts,omitempty"`
	OmittedImages int         `json:"omitted_images,omitempty"`
	LastReplies   []LastReply `json:"last_replies,omitempty"`
}

type LastReply struct {
	No       int    `json:"no"`
	Now      string `json:"now"`
	Name     string `json:"name"`
	Com      string `json:"com"`
	Time     int    `json:"time"`
	Resto    int    `json:"resto"`
	ID       string `json:"id"`
	Filename string `json:"filename,omitempty"`
	Ext      string `json:"ext,omitempty"`
	W        int    `json:"w,omitempty"`
	H        int    `json:"h,omitempty"`
	TnW      int    `json:"tn_w,omitempty"`
	TnH      int    `json:"tn_h,omitempty"`
	Tim      int64  `json:"tim,omitempty"`
	Md5      string `json:"md5,omitempty"`
	Fsize    int    `json:"fsize,omitempty"`
}
