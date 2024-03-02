/*
	=====================================================================

** terminalSize package **
This package tracks the size of the terminal, and adjusts the application
accordingly; basically it makes the application responsive to give a good
user experience on both small and big screens.

    =====================================================================
*/

package terminalSize

// import:
// the only dependency for this package is the golang.org/x/term package
// which is used to do the size handling stuff.

import "github.com/nsf/termbox-go"

/*
    =====================================================================
GetTerminalSize function:
input: none
output: 2 integers (width and height of the terminal)
The function returns the current width and height of the terminal.

*/

func GetTerminalSize() (int, int) {
	// In order to get the width and height of the terminal, we use the term.GetSize function.
	width, height := termbox.Size()
	// Return the width and the height obtained
	return width, height
}

/*
    =====================================================================
GetWordsPerLine function:
input: none
output: int (the number of words per line)
This function calculates the appropriate number of words per line in order
to give a good user experience to the user, since it does so while taking
into account the size of the terminal (in this case we are interested solely
in the width).

*/

func GetWordsPerLine() int {
	// Get the width and ignore the height.
	width, _ := GetTerminalSize()
	// Some if/else clauses involving the width.
	if width < 50 {
		// If the width is less than 50, the number of words per line will be 5, i.e each line in the text reader will have 5 words.
		return 5
	} else if width < 80 {
		// the logic is the same as above
		return 7
	} else if width < 100 {
		// the logic is the same as above
		return 10
	} else {
		// the logic is the same as above
		return 15
	}
}

/*
    =====================================================================
GetLinesPerPage function:
input: none
output: int (the number of lines per page)
This function calculates the appropriate number of lines per page in order
to give a good user experience to the user, since it does so while taking
into account the size of the terminal (in this case we are interested solely
in the height).

*/

func GetLinesPerPage() int {
	// Get current height of the terminal (ignore the width).
	_, height := GetTerminalSize()

	// Some if/else clauses involving the height
	if height < 12 {
		// If the height is less than 12, the number of lines per page will be 5, i.e each page in the text reader will have 5 lines.
		return 5
	} else if height < 20 {
		// all the logic below is analogous to the first case; i.e for a given height range, return a number of lines per page.
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
