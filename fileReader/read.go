/*
	=====================================================================

** fileReader package **
This package is responsible for the tokenization of the texts we are going
to load, as well as the storage of the levels of knowledge of the words in
a particular language when we are studying.
This package is also responsible for the creation of "dictionary" files that
can be exported to anki or memrise.

    =====================================================================
*/

package fileReader

/*
Imported packages:
1) encoding/json --> used to communicate with the locally stored files that store
our levels of knowledge of words, since they are stored in json.
2) fmt --> used to print out stuff to the console in case something goes wrong
3) io/ioutil --> used to work with files
4) log --> used for error handling
5) os --> used to work with files

We are then importing the terminalSize and the translator packages, to use their features.
For more info on them, go in the directory ../terminalSize and ../translator

*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"example.com/packages/terminalSize"
	"example.com/packages/translator"
)

/*
Text struct
This is the struct for the text we will open in our application.
*/

type Text struct {
	TextContent         string         // This is the actual content of the text.
	Length              int            // This is the length of the content of the text.
	TokenList           []string       // This is the list of tokens (or more informally single words) of the text: in european languages, tokens are almost always separated by spaces.
	TokenCursorPosition int            // This is the current position of our cursor (i.e the current word we're hovering in our application)
	TokenLength         int            // This is the total number of tokens (words) in the text.
	Pages               int            // This is the number of pages of the text.
	PageList            [][]string     // This is the list of tokens in the pages of the text.
	CurrentPage         int            // This is the number that displays the current page in which we are in.
	WordLevels          map[string]int // This is the map object that stores the levels of knowledge that we have for a certain word
	// level of knowledge 0 --> ignore
	// 1 --> don't know
	// 2 --> meh
	// 3 --> know well
	CurrentTranslate string // This holds the value of the translation of the word we're currently hovering over (if we requested a translation with the key "5").
}

/*
ReturnFileContent function
input: filename (string), which is the name of the text file we're going to open
output: string; it is the content of the text file we're opening.
*/

func ReturnFileContent(filename string) string {
	// Read the content of the text file, and if there's an error, store it inside the err variable.
	content, err := ioutil.ReadFile(filename)
	// Conver the content variable (which is a slice of bytes) into a string.
	actualContent := string(content)
	// Add a space at the end because otherwise the last character will be skipped.
	actualContent = fmt.Sprintf("%s ", actualContent)
	// Some error handling.
	if err != nil {
		log.Fatal(err)
	}

	// Return the string containing the content of the file.
	return actualContent
}

/*
TokenizeText function:
input: the content of the text (string)
output: a slice of strings, i.e the list of words (tokens) of the text.
This function is the one that tokenizes the text file.
*/

// (Notice that this tokenization only works for languages that work like
// european languages [e.g latin, indonesian, tagalog, russian, serbian, italian
// latin, esperanto exc. ])
// For chinese there is another tokenization function defined later.
// For languages like arabic and japanese it's considerably more difficult.

func TokenizeText(text string) []string {
	// initialize the slice we're going to return
	var output []string
	// Loop through the characters of the string
	i := 0
	// If the character encountered is an empty space, skip it
	for string(text[i]) == " " || string(text[i]) == "\n" || string(text[i]) == "\t" {
		i++
	}
	// If there are no more empty spaces, then we can start scanning for an actual word
	// start scanning word
	for i < len(text)-1 {
		// Declare variable j and initialize it to the current value of i.
		// With this variable j we will store the beginning of the word (token).
		// With i we will reach the end of the current word (current token).
		j := i
		// We keep incrementing i until we hit an empty space (which terminates the current word)
		for string(text[i]) != " " && i < len(text)-1 && string(text[i]) != "\n" && string(text[i]) != "\t" {
			i++
		}
		// We then set the token equal to text[j:i], using the string slicing provided by Go.
		token := text[j:i]
		// We append the scanned token to the slice we're going to return (which we called "output")
		output = append(output, token)
		// Skip other empty spaces
		if i < len(text)-1 {
			for string(text[i]) == " " || string(text[i]) == "\n" || string(text[i]) == "\t" && i < len(text)-1 {
				i++
			}
		}
	}
	// Return our slice of tokens.
	return output
}

/*
DivideInPages function:
input: The list of tokens of the text, which strictly speaking is represented as a slice of strings.
output: a slice of slices of strings (which is basically the list of pages,
then each page is represented as the list of the tokens inside of it.)

The following function is responsible for the determination of the number of pages in which the text opened will
be divided, by taking into account the size of the terminal, making thus the feel of the application responsive.
*/

func DivideInPages(tokens []string) [][]string {
	// To see more info on what the 2 methods used below do, check the comments in the terminalSize.go file inside the terminalSize directory.
	// For now just know that these 2 lines set words and lines equal to the preferred number of words per line and lines per page, by taking into account the size of the terminal.
	words := terminalSize.GetWordsPerLine()
	lines := terminalSize.GetLinesPerPage()
	// The total number of words in a page is thus given by the product of the total number
	// of words per line with the total number of lines per page.
	total := words * lines

	// Store the length of the token list of the text inside a variable called length
	length := len(tokens)
	// The number of pages is the integer quotient of length by total.
	pages := (length / total)
	// Declare the variable we're going to return.
	var outputSlice [][]string
	var endIndex int
	// Form the pages with a loop
	for i := 0; i < pages; i++ {
		startIndex := i * total
		endIndex = (i + 1) * total

		slice := tokens[startIndex:endIndex]
		outputSlice = append(outputSlice, slice)
	}
	// We deal with the last slice separately.
	// (we just get the last tokens left)
	lastSlice := tokens[endIndex:length]
	// We then append it.
	outputSlice = append(outputSlice, lastSlice)
	// Return our slice of slices (which is the list of pages of the text).
	return outputSlice
}

/*
CheckIfContentIsNil function:
input: a text
output: a boolean value
The following function just checks if a text passed in is just comprised of spaces/newline characters/tabs.
*/

func CheckIfContentIsNil(st string) bool {
	// initialize the flag we're returning to true.
	emptyFlag := true
	for _, v := range st {
		// check
		if string(v) != " " && string(v) != "\n" && string(v) != "\t" {
			emptyFlag = false
		}
	}
	return emptyFlag
}

/*
MakeJsonFile function:
inputs: 1) a map (which is how well you know the words in the text)
2) a language.

This function makes the json file that stores the knowledge data about the
words in a particular language.
*/

func MakeJsonFile(data map[string]int, language string) {
	filename := fmt.Sprintf("languages/%s/words.json", language)
	jsonData, err1 := json.Marshal(data)

	if err1 != nil {
		fmt.Println("Error while marshalling the data", err1)
		return
	}

	file, err2 := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err2 != nil {
		fmt.Println("Error opening file:", err2)
		return
	}
	defer file.Close() // Close the file when we're done

	// Write data to the file
	_, err3 := file.Write(jsonData)

	if err3 != nil {
		fmt.Println("Error writing to file:", err3)
		return
	}
}

func LoadJsonWords(filepath string) map[string]int {
	content, err := ioutil.ReadFile(filepath)
	actualContent := string(content)
	var data map[string]int
	err2 := json.Unmarshal([]byte(actualContent), &data)

	if err2 != nil || err != nil {
		fmt.Printf("Error while trying to unmarshal json\n")
	}
	return data
}

/*
FileExists function:
input: the name of the file we're interested in (string)
output: a boolean value (true/false)
This function just checks if a file exists.
*/

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

/*
InitMap function:
inputs:
1) tokens (slice of strings) --> which is the list of tokens of the current text
2) language --> which is the current language we're studying
output: a map, which is the map which represents how well we know the words in the text.
*/

func InitMap(tokens []string, language string) map[string]int {
	fileInQuestion := fmt.Sprintf("languages/%s/words.json", language)
	if FileExists(fileInQuestion) {
		output := LoadJsonWords(fileInQuestion)
		return output
	} else {
		output := make(map[string]int)
		for _, token := range tokens {
			output[token] = 0
		}
		MakeJsonFile(output, language)
		return output
	}
}

func MakeDictFromMenu(language string) map[string]int {
	fileInQuestion := fmt.Sprintf("languages/%s/words.json", language)
	if FileExists(fileInQuestion) {
		output := LoadJsonWords(fileInQuestion)
		return output
	} else {
		return map[string]int{}
	}
}

func MakeDictionary(data map[string]int, language string) {
	filename := fmt.Sprintf("languages/%s/dictionary.txt", language)
	finalString := ""
	finalString += "\n"
	for k, v := range data {
		if v == 1 || v == 2 {
			translation := translator.Translate(k, language)
			finalString += fmt.Sprintf("%s, %s\n", k, translation)
		}
	}
	file, err2 := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err2 != nil {
		fmt.Println("Error opening file:", err2)
		return
	}
	defer file.Close() // Close the file when we're done

	// Write data to the file
	_, err3 := file.Write([]byte(finalString))

	if err3 != nil {
		fmt.Println("Error writing to file:", err3)
		return
	}
}

func InitText(filename string, language string) Text {
	content := ReturnFileContent(filename)
	if !CheckIfContentIsNil(content) {
		var contentLength = len(content)
		var currentCursor int = 0
		TokenList := TokenizeText(content)
		pageList := DivideInPages(TokenList)
		var wordsMap = InitMap(TokenList, language)
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList), WordLevels: wordsMap}
		return outputText
	} else {
		content = "Text file is empty. Are you sure you opened the right one?"
		var contentLength = len(content)
		var currentCursor int = 0
		TokenList := TokenizeText(content)
		pageList := DivideInPages(TokenList)
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList)}
		return outputText
	}
}
