package translator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ResData struct {
	TranslatedText string  `json:"translatedText"`
	Match          float64 `json:"match"`
}
type Response struct {
	ResponseData   ResData `json:"responseData"`
	ResponseStatus int     `json:"responseStatus"`
}

func Translate(text string) string {
	encodedText := url.QueryEscape(text) + "%s"
	apiurl := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=ru|en", encodedText)
	response, err := http.Get(apiurl)
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return ""
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", response.StatusCode)
		return ""
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	var res Response
	err2 := json.Unmarshal(body, &res)

	if err2 != nil {
		fmt.Println("Error parsing JSON:", err2)
		return ""
	}
	if len(res.ResponseData.TranslatedText) >= 2 {
		return res.ResponseData.TranslatedText[:len(res.ResponseData.TranslatedText)-2]
	} else {
		return res.ResponseData.TranslatedText
	}
}
