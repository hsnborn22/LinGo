/*
	=====================================================================

** audioPlayer package **
This package is responsible for the playing of sounds inside the applications,
in particular, the audio files for the words inside the text, via a TTS API.

    =====================================================================
*/

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
	"io/ioutil"
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

func downloadFile(url, filePath string) error {
	// Make the GET request
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code: %d", response.StatusCode)
	}

	// Create the output file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the response body to the output file
	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetAudio(text string, languageId string) {
	url := "https://api.soundoftext.com/sounds"

	// Define the data to be sent in the request body (can be a string or other types)
	data := []byte(fmt.Sprintf(`{"engine": "Google", "data": {"text":"%s", "voice": "%s"}}`, text, languageId))

	// Make the HTTP POST request
	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", response.StatusCode)
		return
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var res Response
	err2 := json.Unmarshal(body, &res)

	if err2 != nil {
		fmt.Println("Error parsing JSON:", err2)
		return
	}

	mp3URL := fmt.Sprintf("https://files.soundoftext.com/%s.mp3", res.Id)

	// Local path where you want to save the downloaded file
	localFilePath := fmt.Sprintf("audio/%s.mp3", text)

	// Call the downloadFile function
	err3 := downloadFile(mp3URL, localFilePath)
	if err3 != nil {
		fmt.Println("Error downloading file:", err3)
		return
	}
}

func PlayMP3(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))

	<-done
	return nil
}

func DeleteMP3(path string) {
	err := os.Remove(path)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}
}
