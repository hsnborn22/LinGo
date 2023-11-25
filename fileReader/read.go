package fileReader

import (
	"io/ioutil"
	"log"
)

type Text struct {
	TextContent         string
	Length              int
	TokenList           []string
	TokenCursorPosition int
	TokenLength         int
}

func ReturnFileContent(filename string) string {
	content, err := ioutil.ReadFile("random.txt")

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

func InitText(filename string) Text {
	content := ReturnFileContent(filename)
	var contentLength = len(content)
	var currentCursor int = 0
	TokenList := TokenizeText(content)
	outputText := Text{TextContent: content, Length: contentLength, TokenList: TokenList, TokenCursorPosition: currentCursor, TokenLength: len(TokenList)}
	return outputText
}
