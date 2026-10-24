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
// We then make a GET request to this url and download the mp3. We play the mp3 with the beep library.
// Downloaded mp3 files are cached inside the audiocache directory (one subdirectory per language),
// so that the same word is only downloaded once.

package audioPlayer

/* Imported packages:
1) bytes --> needed for the manipulation of byte slices
2) encoding/json --> we will need it since the communication with the tts api we're using happens via json
3) fmt --> needed to log possible errors to the console
4) io and io/util --> needed for working with files
5) net/http --> needed to perform the http requests tot he api
6) net/url --> needed to turn words into safe file names for the cache
7) os --> needed for file manipulation
8) path/filepath --> needed to build the paths of the cached files
9) sync --> needed to initialize the speaker only once
10) time --> needed to deal with times
11) beep --> needed to play the mp3

*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
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

// speakerSampleRate is the fixed sample rate the speaker is initialized with.
// Every mp3 is resampled to this rate before being played.
const speakerSampleRate beep.SampleRate = 44100

var (
	speakerOnce    sync.Once
	speakerInitErr error
)

// AudioCacheDir is the directory where the downloaded mp3 files are cached.
const AudioCacheDir = "audiocache"

/*
CachedAudioPath function:
input:
- text (string) = the text we want the audio recording of
- languageId (string) = the id of the language we want the recording in

output: the path of the cached mp3 file for that text and language.
The text is escaped so that words containing characters like "/" still produce a valid file name.
*/

func CachedAudioPath(text string, languageId string) string {
	return filepath.Join(AudioCacheDir, url.PathEscape(languageId), url.PathEscape(text)+".mp3")
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

	// Create a temporary output file in the same directory, so that an interrupted
	// download never leaves a broken mp3 inside the cache.
	out, err2 := os.CreateTemp(filepath.Dir(filePath), "download-*.tmp")
	if err2 != nil {
		return err2.Error()
	}
	tmpPath := out.Name()

	// Copy the response body to the output file
	_, err = io.Copy(out, response.Body)
	closeErr := out.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpPath)
		return err.Error()
	}

	// Move the completed download to its final location
	if err = os.Chmod(tmpPath, 0644); err == nil {
		err = os.Rename(tmpPath, filePath)
	}
	if err != nil {
		os.Remove(tmpPath)
		return err.Error()
	}

	return ""
}

/*
GetAudio function:
input:
- text (string) = the text we want the audio recording of
- languageId (string) = the id of the language we want the recording in

output: 2 strings: the path to the mp3 file and (possibly) an error
The following function gets the mp3 audio for a particular text.
If the audio is already in the cache (see CachedAudioPath) it is returned directly,
otherwise it downloads it from an url given by the API using the downloadFile function
and stores it in the cache.
*/

func GetAudio(text string, languageId string) (string, string) {
	localFilePath := CachedAudioPath(text, languageId)

	// If we already downloaded this word, there is no need to contact the API again
	if info, err := os.Stat(localFilePath); err == nil && info.Size() > 0 {
		return localFilePath, ""
	}

	// Make sure the cache directory for this language exists
	if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
		return "", err.Error()
	}

	apiURL := "https://api.soundoftext.com/sounds"

	// This is the data that will be sent in the request body:
	data := []byte(fmt.Sprintf(`{"engine": "Google", "data": {"text":"%s", "voice": "%s"}}`, text, languageId))

	// Make the HTTP POST request
	response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err.Error()
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		return "", fmt.Sprintf("Unexpected status code: %d\n", response.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Sprintf("Error reading response body: %s", err.Error())
	}

	// Declare a res variable (of type Response defined above)
	// This variable will hold the json that we will receive as a response from the API.
	var res Response
	// unmarshal (i.e parse the json inside the res variable of type Response struct)
	// if there is an error it will be saved inside of the err2 error variable.
	err2 := json.Unmarshal(body, &res)

	// error handling
	if err2 != nil {
		return "", fmt.Sprintf("Error parsins JSON: %s", err2.Error())
	}

	// This is the url to which we will perform the get request to get the mp3 file.
	mp3URL := fmt.Sprintf("https://files.soundoftext.com/%s.mp3", res.Id)

	// Call the downloadFile function
	err3 := downloadFile(mp3URL, localFilePath)
	// some more error handling
	if err3 != "" {
		return "", err3
	}
	return localFilePath, ""
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

	defer streamer.Close()

	// The speaker must be initialized only once: speaker.Init closes and reopens the
	// audio device every time it is called, which crashes the native audio driver.
	speakerOnce.Do(func() {
		speakerInitErr = speaker.Init(speakerSampleRate, speakerSampleRate.N(time.Second/10))
	})
	if speakerInitErr != nil {
		return speakerInitErr.Error()
	}

	done := make(chan bool)
	resampled := beep.Resample(4, format.SampleRate, speakerSampleRate, streamer)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
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
