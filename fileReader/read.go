package fileReader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"example.com/packages/terminalSize"
	"example.com/packages/translator"
)

type Text struct {
	TextContent         string
	Length              int
	TokenList           []string
	TokenCursorPosition int
	TokenLength         int
	Pages               int
	PageList            [][]string
	CurrentPage         int
	WordLevels          map[string]int
	CurrentTranslate    string
}

func ReturnFileContent(filename string) string {
	content, err := ioutil.ReadFile(filename)
	actualContent := string(content)
	actualContent = fmt.Sprintf("%s ", actualContent)

	if err != nil {
		log.Fatal(err)
	}

	return actualContent
}

func TokenizeText(text string) []string {
	// initialize the slice we're going to return
	var output []string
	// Loop through the characters of the string
	i := 0
	for string(text[i]) == " " {
		i++
	}
	// start scanning word
	for i < len(text)-1 {
		j := i
		for string(text[i]) != " " && i < len(text)-1 {
			i++
		}
		token := text[j:i]
		output = append(output, token)
		if i < len(text)-1 {
			for string(text[i]) == " " && i < len(text)-1 {
				i++
			}
		}

	}
	return output
}

func DivideInPages(tokens []string) [][]string {
	words := terminalSize.GetWordsPerLine()
	lines := terminalSize.GetLinesPerPage()
	total := words * lines

	length := len(tokens)
	pages := (length / total)
	var outputSlice [][]string
	var endIndex int
	for i := 0; i < pages; i++ {
		startIndex := i * total
		endIndex = (i + 1) * total

		slice := tokens[startIndex:endIndex]
		outputSlice = append(outputSlice, slice)
	}
	lastSlice := tokens[endIndex:length]
	outputSlice = append(outputSlice, lastSlice)
	fmt.Println(len(outputSlice))
	return outputSlice
}

func CheckIfContentIsNil(st string) bool {
	emptyFlag := true
	for _, v := range st {
		if string(v) != " " {
			emptyFlag = false
		}
	}
	return emptyFlag
}

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

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

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

func MakeDictionary(data map[string]int, language string) {
	filename := "languages/russian/dictionary.txt"
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
