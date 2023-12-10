package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// This is the map that contains how cyrillic letters in all variations of cyrillic
// (russian,mongolian,ukrainian, serbian,kazakh,tajik exc.) are mapped to latin.

var cyrillicToLatin map[string]string = map[string]string{"А": "A", "Б": "B", "В": "V", "Г": "G", "Д": "D", "Ђ": "Đ", "Е": "E", "Ё": "YO", "Ж": "ZH", "З": "Z", "И": "I", "Й": "Y", "К": "K", "Л": "L", "Љ": "LJ", "М": "M", "Н": "N", "Њ": "NJ", "О": "O", "П": "P", "Р": "R", "С": "S", "Т": "T", "У": "U", "Ф": "F", "Х": "KH", "Ц": "TS", "Ч": "CH", "Ш": "SH", "Щ": "SHCH", "Ъ": "''", "Ы": "Y", "Ь": "'", "Э": "E", "Ю": "YU", "Я": "YA",
	"а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "ђ": "đ", "е": "e", "ё": "yo", "ж": "zh", "з": "z", "и": "i", "й": "y", "к": "k", "л": "l", "љ": "lj", "м": "m", "н": "n", "њ": "nj", "о": "o", "п": "p", "р": "r", "с": "s", "т": "t", "у": "u", "ф": "f", "х": "kh", "ц": "ts", "ч": "ch", "ш": "sh", "щ": "shch", "ъ": "''", "ы": "y", "ь": "'", "э": "e", "ю": "yu", "я": "ya",
	"Ң": "NG", "Ү": "U", "Ұ": "U", "Һ": "H", "Ө": "O", "ү": "u", "ұ": "u", "һ": "h", "ө": "o",
	"Ә": "A", "Ғ": "G", "Қ": "Q", "ә": "a", "ғ": "g", "қ": "q", "ң": "n",
}

// This is the map that contains how greek characters are mapped phonetically to
// latin characters.

var greekToLatin = map[string]string{
	"Α": "A",
	"Β": "B",
	"Γ": "G",
	"Δ": "D",
	"Ε": "E",
	"Ζ": "Z",
	"Η": "H",
	"Θ": "TH",
	"Ι": "I",
	"Κ": "K",
	"Λ": "L",
	"Μ": "M",
	"Ν": "N",
	"Ξ": "X",
	"Ο": "O",
	"Π": "P",
	"Ρ": "R",
	"Σ": "S",
	"Τ": "T",
	"Υ": "U",
	"Φ": "PH",
	"Χ": "CH",
	"Ψ": "PS",
	"Ω": "O",
	"α": "a",
	"β": "b",
	"γ": "g",
	"δ": "d",
	"ε": "e",
	"ζ": "z",
	"η": "h",
	"θ": "th",
	"ι": "i",
	"κ": "k",
	"λ": "l",
	"μ": "m",
	"ν": "n",
	"ξ": "x",
	"ο": "o",
	"π": "p",
	"ρ": "r",
	"σ": "s",
	"τ": "t",
	"υ": "u",
	"φ": "ph",
	"χ": "ch",
	"ψ": "ps",
	"ω": "o",
}

// The following functions are responsible for latinizations of various scripts; it is specified in the name.

func LatinizeCyrillic(text string) string {
	var outputString string

	for _, char := range text {
		value, found := cyrillicToLatin[string(char)]
		if found {
			outputString += value
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeGreek(text string) string {
	var outputString string

	for _, char := range text {
		value, found := greekToLatin[string(char)]
		if found {
			outputString += value
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeChinese(text string, data map[string][]string) string {
	var outputString string
	for _, char := range text {
		value, found := data[string(char)]
		if found {
			outputString += value[0]
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeJapanese(text string) string {
	var output string
	url := "https://japonesbasico.com/furigana/procesa.php"
	// This is the data that will be sent in the request body:
	data := []byte(fmt.Sprintf(`{"conversion":"romaji", "japaneseText":"%s", "lang":"en"}`, text))

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
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err.Error()
	}

	// Find and print the value inside the second <td> element
	output = findSecondTdValue(doc)
	return output
}

func findSecondTdValue(n *html.Node) string {
	var result string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "td" {
			// Check if it is the second <td> element
			if node.NextSibling == nil {
				// Get the text content of the second <td>
				result = getTextContent(node.FirstChild)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return result
}

func getTextContent(n *html.Node) string {
	var result string

	if n != nil {
		if n.Type == html.TextNode {
			result += n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result += getTextContent(c)
		}
	}

	return result
}

/* InitHanzi function:
This function is responsible for the "translation" of the data inside the hanzi.json file (i.e the map
that associates chinese characters to their pinyin romanizations) inside a Go map object.
It returns the map we obtain.
*/

func InitHanzi() map[string][]string {
	// Get the raw content of the hanzi.json file.
	jsonData, _ := os.ReadFile("translator/hanzi.json")

	// Initialize the map we're going to return
	var data map[string][]string

	// Use json.Unmarshal to decode the json into the map
	err := json.Unmarshal([]byte(jsonData), &data)
	// If there's an error print it out to the console.
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
	// return the data map.
	return data
}

/* LatinizeText function
This function is responsible for the latinization of a certain portion of a text.
Input: 1) the text we want to latinize (string)
2) data, which is the map that contains the latinizations in pinyin of the chinese hanzi characters.
3) language; which is the language we want to latinize (i.e the language we're currently studying).
*/

func LatinizeText(text string, data map[string][]string, language string) string {
	// Check what language we're studying
	switch language {
	case "chinese":
		// If it's chinese latinize the chinese into pinyin
		return LatinizeChinese(text, data)
	case "russian", "serbian", "mongolian", "belarusian", "ukrainian", "bulgarian", "kazakh":
		// If it's one of these languages, then the script used is cyrillic: latinize accordingly.
		return LatinizeCyrillic(text)
	case "greek":
		// Greek script latinization
		return LatinizeGreek(text)
	case "japanese":
		return LatinizeJapanese(text)
	case "korean":
		return LatinizeKorean(text)
	default:
		// else just returns the text; I still need to cover a lot of non-latin scripts.
		return text

	}
}
