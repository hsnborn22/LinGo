package terminalSize

import "golang.org/x/term"

func GetTerminalSize() (int, int) {
	width, height, err := term.GetSize(0)
	if err != nil {
		return -1, -1
	}
	return width, height
}

func GetWordsPerLine() int {
	width, _ := GetTerminalSize()
	if width < 50 {
		return 5
	} else if width < 80 {
		return 7
	} else if width < 100 {
		return 10
	} else {
		return 15
	}
}
