/*
	=====================================================================

** audioPlayer package **
This package is responsible for the playing of sounds inside the applications,
in particular, the audio files for the words inside the text, via a TTS API.

    =====================================================================
*/

// The API used (api.soundoftext.com) works as follows:
// We first make a POST request, containing the text we want to hear and the language selected.
// The API then responds to us with json containing the url to the mp3 file generated.
// We then make a GET request to this url and download the mp3. We play the mp3 with the beep library,
// and then delete it.

package audioPlayer

/* Imported packages:
1) bytes --> needed for the manipulation of byte slices
2) encoding/json --> we will need it since the communication with the tts api we're using happens via json
3) fmt --> needed to log possible errors to the console
4) io and io/util --> needed for working with files
5) net/http --> needed to perform the http requests tot he api
6) os --> needed for file manipulation
7) time --> needed to deal with times
8) beep --> needed to play the mp3

*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Response struct:
// a struct type that mirrors the json for unmarshalling

type Response struct {
	Success bool   `json:"success"`
	Id      string `json:"id"`
}

/*
downloadFile function:
input: 2 strings: url and filePath
output: 1 (possible) error
This function, as the name suggests, is responsible for the download of
the mp3 file from the api.
*/

func downloadFile(url, filePath string) string {
	// Make the GET request
	response, err := http.Get(url)
	if err != nil {
		return err.Error()
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("Unexpected status code: %d", response.StatusCode)
	}

	// Create the output file
	out, err2 := os.Create(filePath)
	if err2 != nil {
		return err2.Error()
	}
	defer out.Close()

	// Copy the response body to the output file
	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err.Error()
	}

	return ""
}

/*
GetAudio function:
input:
- text (string) = the text we want the audio recording of
- languageId (string) = the id of the language we want the recording in

output: void
The following function gets the mp3 audio for a particular text.
In order to get the audio, it downloads it from an url given by the API
using the downloadFile function.
*/

func GetAudio(text string, languageId string) string {
	url := "https://api.soundoftext.com/sounds"

	// This is the data that will be sent in the request body:
	data := []byte(fmt.Sprintf(`{"engine": "Google", "data": {"text":"%s", "voice": "%s"}}`, text, languageId))

	// Make the HTTP POST request
	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err.Error()
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("Unexpected status code: %d\n", response.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response body: %s", err.Error())
	}

	// Declare a res variable (of type Response defined above)
	// This variable will hold the json that we will receive as a response from the API.
	var res Response
	// unmarshal (i.e parse the json inside the res variable of type Response struct)
	// if there is an error it will be saved inside of the err2 error variable.
	err2 := json.Unmarshal(body, &res)

	// error handling
	if err2 != nil {
		return fmt.Sprintf("Error parsins JSON: %s", err2.Error())
	}

	// This is the url to which we will perform the get request to get the mp3 file.
	mp3URL := fmt.Sprintf("https://files.soundoftext.com/%s.mp3", res.Id)

	// Local path where you want to save the downloaded file
	localFilePath := fmt.Sprintf("audio/%s.mp3", text)

	// Call the downloadFile function
	err3 := downloadFile(mp3URL, localFilePath)
	// some more error handling
	if err3 != "" {
		return err3
	}
	return ""
}

/*
PlayMP3 function:
input: filePath (string) which is the path to the mp3 file we want to play.
output: (possibly) an error
The PlayMP3 function, like the name suggests, is responsible for the sounds played in the
application. It will play the mp3 files using concurrency.
*/

func PlayMP3(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return err.Error()
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err.Error()
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))

	<-done
	return ""
}

/*
DeleteMP3 function
input: path (string), which is the filepath to the mp3 file we want to delete.
output: void (nothing)
This function, like the name suggests, deletes an mp3 file in a specific location.
*/

func DeleteMP3(path string) string {
	// Remove the file; if there is an error, store it inside the err variable
	err := os.Remove(path)
	// Error handling:
	// if there is an error, we get a notification.
	if err != nil {
		return err.Error()
	}
	return ""
}
