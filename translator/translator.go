/*
	=====================================================================

** translator package **
This package is responsible for the translation of words in the text we
are going to study. It does so by communicating with an API through http
requests (the mymemory translation API).

    =====================================================================
*/

package translator

/* Imports:
1) encoding/json --> used to parse the json we receive via GET requests from the API, as well as
transform some of the data of our application into JSON format for using it for POST requests.
2) fmt --> used to print out to the console possible errors
3) io/ioutil --> used to read files.
4) net/http --> used to make the http requests to the API.
5) net/url --> used to transform the words of languages which do not use the latin alphabet (like russian,
ukrainian, mongolian,kazakh, standard arabic, korean, chinese, greek exc.) into url encoded format.

*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

/*
ResData, Match and Response structs:

The following three structs mirror the json structure of the
responses of the API, and we use them to unmarshall the json responses
of the translate API into golang structs that are use-able in our
application.

*/

type ResData struct {
	// The translated text
	TranslatedText string `json:"translatedText"`
	// Match --> indicates how accurate the translation is
	Match float64 `json:"match"`
}

type Match struct {
	// The original text (i.e the text in the original language)
	Segment string `json:"segment"`
	// The translated text
	Translation string `json:"translation"`
	// The source language (i.e the language it is being translated from)
	Source string `json:"source"`
	// The target language
	Target string `json:"target"`
	// The accuracy of the translation, represented by a float.
	Match float64 `json:"match"`
}

type Response struct {
	// ResponseData field contains the content of the "main" translation
	// (i.e the one that the algorithm deems to be the most accurate)
	ResponseData ResData `json:"responseData"`
	// ResponseStatus is the status of the response message (i.e 200 OK)
	ResponseStatus int `json:"responseStatus"`
	// You can ignore these 2 fields
	QuotaFinished   bool   `json:"quotaFinished"`
	ResponseDetails string `json:"responseDetails"`
	// Matches contains the other translations of the word.
	Matches []Match `json:"matches"`
}

/*
    =====================================================================
Translate function:
input: string (the text in the original language)
output: string (the translated text)
This function takes in a text in the source language and returns a translation in
the target language.
*/

func Translate(text string) string {
	// This piece of code encodes the text in url encoding (since it might possibly be a not valid url)
	encodedText := url.QueryEscape(text) + "%s"
	// The url to which we will perform the get request.
	apiurl := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=ru|en", encodedText)
	// Response and error object originating from the request to the above url.
	response, err := http.Get(apiurl)
	// If there is an error, print it out to the console.
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return ""
	}
	// We are going to close the response.Body eventually and we defer it here to the end.
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", response.StatusCode)
		return ""
	}

	// Read the response body
	body, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		fmt.Println("Error reading response body:", err2)
		return ""
	}

	// Initialize a variable res of type Response, to store in the output json as a struct.
	var res Response
	// Parse the json we received as a response
	err3 := json.Unmarshal(body, &res)

	if err3 != nil {
		fmt.Println("Error parsing JSON:", err3)
		return ""
	}

	// Return the translation: we're doing a little case-checking to adjust the output
	// for some problems regarding text encoding (since we might potentially have non-latin
	// alphabet languages)
	if len(res.ResponseData.TranslatedText) >= 2 {
		return res.ResponseData.TranslatedText[:len(res.ResponseData.TranslatedText)-2]
	} else {
		return res.ResponseData.TranslatedText
	}
}
