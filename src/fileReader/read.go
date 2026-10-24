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
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"example.com/packages/languageHandler"
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
	WordList            []string       // This is the normalized version of every token in TokenList (same positions), see NormalizeWord. It's the key used for the word levels, translations and audio.
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
	CurrentTranslate    string // This holds the value of the translation of the word we're currently hovering over (if we requested a translation with the key "5").
	CurrentLatinization string
}

// Used to communicate with the japanese nlp tokenization python script.

type MyData struct {
	Tokens []string `json:"tokens"`
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
		for (string(text[i]) != " " && string(text[i]) != "\n" && string(text[i]) != "\t") && i < len(text)-1 {
			i++
		}
		// We then set the token equal to text[j:i], using the string slicing provided by Go.
		token := text[j:i]
		// We append the scanned token to the slice we're going to return (which we called "output")
		output = append(output, token)
		// Skip other empty spaces
		if i < len(text)-1 {
			for (string(text[i]) == " " || string(text[i]) == "\n" || string(text[i]) == "\t") && i < len(text)-1 {
				i++
			}
		}
	}
	// Return our slice of tokens.
	return output
}

/*
NormalizeWord function:
inputs: a token (string) and the language we're studying (string)
output: the normalized word (string)

Tokens keep the punctuation attached to them ("Mutta," "talo." "¿Qué"), and the same word can appear
with or without a capital letter at the start of a sentence. This function returns the form of the word
that we use as a key (for the word levels, the translations, the audio exc.), so that all these variants
are treated as the same word. The token itself is left untouched, so the text is still displayed as it is.

1) The punctuation marks and symbols at the beginning and at the end of the token are removed.
We use the unicode categories instead of a fixed list of characters, so that the punctuation of every
language is recognised (e.g ¿ ¡ « » „ “ 、 。 ، ؟ । ...). Punctuation inside the word is kept,
since it's often part of the word (l'homme, well-known, aujourd'hui).
2) The word is lowercased, using the turkish rules for turkish and azerbaijani (I --> ı, İ --> i).
Scripts without cases (chinese, arabic, thai exc.) are not affected.

If the token is only made of punctuation (e.g "—" or "..."), an empty string is returned:
it is not a word, so it can't have a level of knowledge.
*/

func NormalizeWord(token string, language string) string {
	word := strings.TrimFunc(token, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r)
	})
	if language == "turkish" || language == "azerbaijani" {
		return strings.ToLowerSpecial(unicode.TurkishCase, word)
	}
	return strings.ToLower(word)
}

/*
NormalizeTokens function:
It returns the normalized version (see NormalizeWord) of every token in the list.
*/

func NormalizeTokens(tokens []string, language string) []string {
	output := make([]string, len(tokens))
	for i, token := range tokens {
		output[i] = NormalizeWord(token, language)
	}
	return output
}

/*
CurrentWord method:
It returns the normalized word the cursor is currently on.
*/

func (t Text) CurrentWord() string {
	return t.WordList[t.TokenCursorPosition]
}

/*
SetCurrentWordLevel method:
It sets the level of knowledge of the word the cursor is on and saves it in the json file.
Tokens that are only punctuation are skipped.
*/

func (t Text) SetCurrentWordLevel(level int, language string) {
	word := t.CurrentWord()
	if word == "" {
		return
	}
	t.WordLevels[word] = level
	MakeJsonFile(t.WordLevels, language)
}

/*
TokenizeChineseText function:

This function provides a tokenization for chinese (both simplified and traditional) texts.

*/

func TokenizeChineseText(text string) []string {
	// initialize the slice we're going to return
	var output []string
	// chineseString, _ := utf8.DecodeRuneInString(text)
	// Loop through the characters of the string
	// If there are no more empty spaces, then we can start scanning for an actual pictogram
	// In chinese, we will denote each pictogram as a token.
	// start scanning
	for _, char := range text {
		if string(char) != " " && string(char) != "\n" && string(char) != "\t" {
			output = append(output, string(char))
		}
	}
	// Return our slice of tokens.
	return output
}

/*
TokenizeJapaneseText function:

This function provides a tokenization for japanese text (which is not a language that employs spaces to separate words).
*/

func TokenizeJapaneseText(text string) []string {
	serverAddr := "127.0.0.1:8080"

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return []string{}
	}
	defer conn.Close()

	stripped := strings.ReplaceAll(text, "\t", "")
	stripped = strings.ReplaceAll(stripped, " ", "")
	stripped = strings.ReplaceAll(stripped, "\n", "")

	conn.Write([]byte(stripped))

	buffer := make([]byte, 262144)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return []string{err.Error()}
	}

	receivedData := string(buffer[:bytesRead])
	// Create an instance of your Go struct
	var myData MyData

	// Unmarshal the JSON data into the struct
	err2 := json.Unmarshal([]byte(receivedData), &myData)
	if err2 != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return []string{err2.Error()}
	}
	fmt.Println(myData.Tokens)
	return myData.Tokens
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

/*
LoadJsonWords function:

input: filepath (string) --> the location in memory of our json file storing the
word levels.
output: map --> which is the levels of knowledge of words represented as a Go map.

The following function is responsible for loading the levels of knowledge for various words
starting from a json file storing them into a map object.
This function is used in the InitMap method, which is very important for the logic of the program.
*/

func LoadJsonWords(filepath string) map[string]int {
	// Read the content of the file (which is Json).
	// if there is an error, save it in the err variable
	content, err := ioutil.ReadFile(filepath)
	// Convert the content we got to string type
	// in order to work more easily with it throughout the application
	actualContent := string(content)
	// Initialize our return value (our map)
	var data map[string]int
	// Unmarshal (i.e parse the json) into the Go data map variable data.
	err2 := json.Unmarshal([]byte(actualContent), &data)

	// If there is an error, tell us.
	if err2 != nil || err != nil {
		fmt.Printf("Error while trying to unmarshal json\n")
	}
	// return the map.
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
1) words (slice of strings) --> which is the list of normalized words of the current text (see NormalizeWord)
2) language --> which is the current language we're studying
output: a map, which is the map which represents how well we know the words in the text.

This function is responsible for the creation of a map file storing the levels of knowledge of the various words we encounter.
If there is an existing json files that already has some levels of knowledge for words saved, import that.
If not, create it.
*/

func InitMap(words []string, language string) map[string]int {
	// This is the location of our words.json file, which stores on our disk (non-volatile memory)
	// the levels of knowledge of various words for the language we're studying.
	fileInQuestion := fmt.Sprintf("languages/%s/words.json", language)
	output := make(map[string]int)
	// If the file exists, load the levels from the file.
	if FileExists(fileInQuestion) {
		// Older versions of the app saved the tokens as they were (with punctuation and capital letters),
		// so we normalize the keys again: if two of them become the same word, keep the highest level.
		changed := false
		for k, v := range LoadJsonWords(fileInQuestion) {
			word := NormalizeWord(k, language)
			if word != k {
				changed = true
			}
			if word == "" {
				continue
			}
			if old, ok := output[word]; !ok || v > old {
				output[word] = v
			}
		}
		if changed {
			MakeJsonFile(output, language)
		}
		return output
	}
	// If not, create it and initialize all the knowledge levels for the words in the text to 0 (ignore).
	for _, word := range words {
		if word != "" {
			output[word] = 0
		}
	}
	MakeJsonFile(output, language)
	return output
}

/*
MakeDictFromMenu function
input: language (string)
output: it returns a map, which contains as keys the words and as values the levels of knowledge.
This function allows you to create a dictionary file (i.e a file containing pairs of words in source-target language).
This file is formatted in such a way that you can import it in both memrise and anki and study the pairs you encountered
as flash cards.
The creation of the actual file is done in the main file, so this function is just an intermediary. In fact, it returns a
map as you can see. From this map we then quickly create the file we cited above.
*/

func MakeDictFromMenu(language string) map[string]int {
	fileInQuestion := fmt.Sprintf("languages/%s/words.json", language)
	if FileExists(fileInQuestion) {
		output := LoadJsonWords(fileInQuestion)
		return output
	} else {
		return map[string]int{}
	}
}

/*
MakeDictionary function

This function is responsible for the actual creation of the dictionary file that can then
be exported to Anki,memrise and other flashcard systems. It takes in a map[string]int which represents
the levels of knowledge of determinate words, and the target language we're studying.
With these informations it then creates a file that contains couplets of the form:

<word in language we're studying>, <word in language we understand>

Example:

hola, hello
gracias, thanks
como estas, how are you

These files can then be exported and made into flashcards using Anki or memrise.
*/

func MakeDictionary(data map[string]int, language string, bootLanguage string) string {
	languageId := languageHandler.LanguageMap2[language]
	return writeDictionary(data, language, func(word string) string {
		// get the translation via the API
		translation, _ := translator.Translate2(word, languageId, bootLanguage)
		return fmt.Sprintf("%s, %s\n", word, translation)
	})
}

/*
MakeAltDictionary function

It differs from the MakeDictionary function since this makes flashcards with latinization included too
This function is responsible for the actual creation of the dictionary file that can then
be exported to Anki,memrise and other flashcard systems. It takes in a map[string]int which represents
the levels of knowledge of determinate words, and the target language we're studying.
With these informations it then creates a file that contains couplets of the form:

<word in language we're studying>, <word in language we understand> (<latinization>)

Example:

привет, hello (privet)
...

These files can then be exported and made into flashcards using Anki or memrise.
*/

func MakeAltDictionary(data map[string]int, language string, bootLanguage string, hanziData map[string][]string) string {
	languageId := languageHandler.LanguageMap2[language]
	return writeDictionary(data, language, func(word string) string {
		// get the translation and the latinization via the APIs
		translation, _ := translator.Translate2(word, languageId, bootLanguage)
		latinization := translator.LatinizeText(word, hanziData, language)
		return fmt.Sprintf("%s, %s (%s)\n", word, translation, latinization)
	})
}

/*
writeDictionary function

Shared by MakeDictionary and MakeAltDictionary: it builds the line of every word we don't know
(i.e. with code 1 or 2) using makeLine, and saves all of them in languages/<language>/dictionary.txt.
makeLine calls the APIs, so the lines are built by a few goroutines at the same time instead of one after
the other, which makes big dictionaries much faster to create.
It returns an error string ("" if everything went fine) instead of printing it, since printing would
mess up the interface.
*/

func writeDictionary(data map[string]int, language string, makeLine func(word string) string) string {
	// Path where we will save our dictionary file.
	filename := fmt.Sprintf("languages/%s/dictionary.txt", language)

	words := []string{}
	for k, v := range data {
		if v == 1 || v == 2 {
			words = append(words, k)
		}
	}
	sort.Strings(words)

	// Each goroutine takes the indexes of the words from the channel and saves the line in the same position.
	const workers = 8
	lines := make([]string, len(words))
	indexes := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range indexes {
				lines[i] = makeLine(words[i])
			}
		}()
	}
	for i := range words {
		indexes <- i
	}
	close(indexes)
	wg.Wait()

	finalString := "\n" + strings.Join(lines, "")

	// os.WriteFile creates the file if it doesn't exist, and overwrites it otherwise.
	if err := os.WriteFile(filename, []byte(finalString), 0666); err != nil {
		return "Error writing the dictionary file: " + err.Error()
	}
	return ""
}

/*
InitText function:
input: the name of the file we opened(string) and the current language we're studying (a string).
output: a Text object

The following function is responsible for the creation (from this the name InitText)
of a Text struct, which represents the current text opened in the application.
*/

func InitText(filename string, language string) Text {
	// Get the content inside the file as a string.
	content := ReturnFileContent(filename)
	// Initialize the cursor position to 0 (i.e to the start).
	var currentCursor int = 0
	// if the file has some characters which are not empty spaces, tabs or new lines
	// and is not chinese, then do the following
	if !CheckIfContentIsNil(content) && language != "chinese" && language != "japanese" && language != "thai" && language != "khmer" && language != "burmese" && language != "lao" {
		// Calculate the length of the content inside the file.
		var contentLength = len(content)
		// Tokenize the text (i.e split it in tokens) using the TokenizeText function
		TokenList := TokenizeText(content)
		// Get the list of pages; each page will be a list of tokens.
		pageList := DivideInPages(TokenList)
		// initialize a word map using the InitMap function: this denotes the level of knowledge of the words inside the text
		// in a particular language.
		WordList := NormalizeTokens(TokenList, language)
		var wordsMap = InitMap(WordList, language)
		// Create outputText object
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, WordList: WordList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList), WordLevels: wordsMap}
		// Return it.
		return outputText
	} else if !CheckIfContentIsNil(content) && language == "chinese" {
		// Calculate the length of the content inside the file.
		// In this case, since we're dealing with chinese, we will have to use the utf8.RuneCountInString method instead.
		// This will actually count the number of characters, in contrast to the "len" function which will just return the byte length.
		var contentLength = utf8.RuneCountInString(content)
		TokenList := TokenizeJapaneseText(content)
		pageList := DivideInPages(TokenList)
		WordList := NormalizeTokens(TokenList, language)
		var wordsMap = InitMap(WordList, language)
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, WordList: WordList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList), WordLevels: wordsMap}
		return outputText
	} else if !CheckIfContentIsNil(content) && (language == "japanese" || language == "khmer" || language == "thai" || language == "burmese" || language == "lao") {
		// Calculate the length of the content inside the file.
		// In this case, since we're dealing with chinese, we will have to use the utf8.RuneCountInString method instead.
		// This will actually count the number of characters, in contrast to the "len" function which will just return the byte length.
		var contentLength = utf8.RuneCountInString(content)
		TokenList := TokenizeJapaneseText(content)
		pageList := DivideInPages(TokenList)
		WordList := NormalizeTokens(TokenList, language)
		var wordsMap = InitMap(WordList, language)
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, WordList: WordList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList), WordLevels: wordsMap}
		return outputText
	} else {
		content = "Text file is empty. Are you sure you opened the right one?"
		// Calculate the length of the content inside the file.
		var contentLength = len(content)
		TokenList := TokenizeText(content)
		pageList := DivideInPages(TokenList)
		outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, WordList: NormalizeTokens(TokenList, language), TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList)}
		return outputText
	}
}
