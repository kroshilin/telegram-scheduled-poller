package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type picturer struct {
	Client *http.Client
	BaseUrl string
}

type searchResponse struct {
	Data randomPicture `json:"data"`
}

type randomPicture struct {
	Image image  `json:"randomPicture"`
}

type image struct {
	ContentUrl string `json:"contentUrl"`
}

func (p picturer) GiveMePictureOf(query string) string {
	requestBody, _ := json.Marshal(map[string]string{
		"query": "{ randomPicture(tag:\"" + query + "\") {contentUrl} }",
		"operationName": "",
		"variables": "",
	});
	req, err := http.NewRequest("POST", p.BaseUrl + "/graphql/", bytes.NewBuffer(requestBody))
	log.Println(p.BaseUrl)
	if err != nil {
		// handle error
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		// handle error
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	data := searchResponse{}
	json.Unmarshal(body, &data)
	return data.Data.Image.ContentUrl
}
