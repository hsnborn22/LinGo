package fileReader

import (
	"fmt"
	"io/ioutil"
	"log"

	"example.com/packages/terminalSize"
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
			for string(text[i]) == " " {
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

func InitMap(tokens []string) map[string]int {
	output := make(map[string]int)
	for _, token := range tokens {
		output[token] = 1
	}
	return output
}

func InitText(filename string) Text {
	content := ReturnFileContent(filename)
	if !CheckIfContentIsNil(content) {
		var contentLength = len(content)
		var currentCursor int = 0
		TokenList := TokenizeText(content)
		pageList := DivideInPages(TokenList)
		var wordsMap = InitMap(TokenList)
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
