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

func GetLinesPerPage() int {
	_, height := GetTerminalSize()

	if height < 12 {
		return 5
	} else if height < 20 {
		return 8
	} else if height < 25 {
		return 12
	} else if height < 30 {
		return 15
	} else if height < 40 {
		return 20
	} else if height < 50 {
		return 25
	} else {
		return 28
	}
}
