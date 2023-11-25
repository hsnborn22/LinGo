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
}

func ReturnFileContent(filename string) string {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}
	return string(content)
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
		for string(text[i]) == " " {
			i++
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

func InitText(filename string) Text {
	content := ReturnFileContent(filename)
	var contentLength = len(content)
	var currentCursor int = 0
	TokenList := TokenizeText(content)
	pageList := DivideInPages(TokenList)
	outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList), CurrentPage: 0, PageList: pageList, Pages: len(pageList)}
	return outputText
}
