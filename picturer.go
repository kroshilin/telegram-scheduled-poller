package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type picturer struct {
	Login, Password string
	Client *http.Client
}

type searchResponse struct {
	Page   int      `json:"page"`
	Data []image `json:"data"`
}

type image struct {
	Assets imageQualities `json:"assets"`
}

type imageQualities struct {
	Thumb thumb `json:"huge_thumb"`
}

type thumb struct {
	Url string `json:"url"`
}

/** Accepts search string and returns link to picture from shutterstock search **/
func (p picturer) GiveMePictureOf(query string) string {
	req, err := http.NewRequest("GET", "https://api.shutterstock.com/v2/images/search?orientation=horizontal&people_age=20s&sort=random&image_type=photo&query=" + query, nil)
	if err != nil {
		// handle error
	}
	req.SetBasicAuth(p.Login, p.Password)
	resp, err := p.Client.Do(req)
	if err != nil {
		// handle error
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	data := searchResponse{}
	json.Unmarshal(body, &data)
	img := strings.Replace(data.Data[0].Assets.Thumb.Url, "image-photo", "z", 1)
	img = strings.Replace(img, "260nw-", "", 1)
	return img
}