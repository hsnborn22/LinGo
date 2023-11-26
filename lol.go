package main

import (
	"fmt"
	"os"
	"path/filepath"

	"example.com/packages/fileReader"
	"example.com/packages/terminalSize"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the app

var (
	titleStyle         = lipgloss.NewStyle().MarginLeft(2)
	itemStyle          = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	quitTextStyle      = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	notKnownItemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	semiKnownItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCA3A"))
	knownItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b300"))
)

func visitFile(fp string, fi os.DirEntry, err error) error {
	if err != nil {
		fmt.Println(err) // can't walk here,
		return nil       // but continue walking elsewhere
	}
	if fi.IsDir() {
		return nil // not a file. ignore.
	}
	// Append the file path to the slice
	filePaths = append(filePaths, fp)
	return nil
}

var filePaths []string // Declare a global slice to store file paths

type model struct {
	choices        []string // items on the to-do list
	cursor         int      // which to-do list item our cursor is pointing at
	viewIndex      int      // viewIndex --> will be 0 for the menu, and 1 for an opened file.
	openedFile     string   // will store the name of the file we opened.
	openedFileText fileReader.Text
}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		choices:   filePaths,
		viewIndex: 0,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.viewIndex {
	case 0:
		switch msg := msg.(type) {

		// Is it a key press?
		case tea.KeyMsg:

			// Cool, what was the actual key pressed?
			switch msg.String() {

			// These keys should exit the program.
			case "ctrl+c", "q":
				return m, tea.Quit

			// The "up" and "k" keys move the cursor up
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			// The "down" and "j" keys move the cursor down
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "enter", " ":
				m.viewIndex = 1
				m.openedFile = m.choices[m.cursor]
				text := fileReader.InitText(m.openedFile)
				m.openedFileText = text
			}
		}

		// Return the updated model to the Bubble Tea runtime for processing.
		// Note that we're not returning a command.
	case 1:
		switch msg := msg.(type) {

		// Is it a key press?
		case tea.KeyMsg:

			// Cool, what was the actual key pressed?
			switch msg.String() {

			// These keys should exit the program.
			case "ctrl+c", "q":
				return m, tea.Quit

			// The "up" and "k" keys move the cursor up
			case "left", "h":
				if m.openedFileText.TokenCursorPosition > 0 {
					m.openedFileText.TokenCursorPosition--
				}

			// The "down" and "j" keys move the cursor down
			case "right", "l":
				if m.openedFileText.TokenCursorPosition < m.openedFileText.TokenLength-1 {
					m.openedFileText.TokenCursorPosition++
				}

			case "up", "k":
				line := terminalSize.GetWordsPerLine()
				if m.openedFileText.TokenCursorPosition > 0 && m.openedFileText.TokenCursorPosition-line >= 0 {
					m.openedFileText.TokenCursorPosition -= line
				}
			case "down", "j":
				line := terminalSize.GetWordsPerLine()
				if m.openedFileText.TokenCursorPosition < m.openedFileText.TokenLength-1 && m.openedFileText.TokenCursorPosition+line < m.openedFileText.TokenLength-1 {
					m.openedFileText.TokenCursorPosition += line
				}
			case "d":
				if m.openedFileText.CurrentPage < m.openedFileText.Pages-1 {
					m.openedFileText.CurrentPage++
				}
			case "a":
				if m.openedFileText.CurrentPage > 0 {
					m.openedFileText.CurrentPage--
				}

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "b":
				m.viewIndex = 0
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var s string
	if m.viewIndex == 0 {
		// The header
		s = "What text file do you want to open?\n\n"

		// Iterate over our choices

		for i, choice := range m.choices {

			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}

			// Is this choice selected?
			var checked = "x" // selected!

			// Render the row
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}

		// The footer
		s += "\nPress q to quit.\n"
	} else if m.viewIndex == 1 {
		wordsPerLine := terminalSize.GetWordsPerLine()
		linesPerPage := terminalSize.GetLinesPerPage()
		width, height := terminalSize.GetTerminalSize()

		s = "Hello you are in " + m.openedFile + " and cursor is at: "
		s += fmt.Sprintf("%v", m.openedFileText.TokenCursorPosition)
		s += fmt.Sprintf("\nCurrent size: %v %v\n", width, height)
		s += "\n"
		for k, element := range m.openedFileText.PageList[m.openedFileText.CurrentPage] {
			var padding1 string = ""
			var padding2 string = " "
			if k%wordsPerLine == 0 && k != 0 {
				padding2 = "\n"
			}
			s += padding1
			actualKey := k + (m.openedFileText.CurrentPage * wordsPerLine * linesPerPage)
			if actualKey == m.openedFileText.TokenCursorPosition {
				s += selectedItemStyle.Render(element)
			} else if value, ok := m.openedFileText.WordLevels[element]; ok {
				switch value {
				case 0:
					s += element
				case 1:
					s += notKnownItemStyle.Render(element)
				case 2:
					s += semiKnownItemStyle.Render(element)
				case 3:
					s += knownItemStyle.Render(element)
				default:
					s += element
				}
			} else {
				s += element
			}
			s += padding2
		}
		s += "\n"
		s += fmt.Sprintf("%v", m.openedFileText.TokenCursorPosition)
		s += fmt.Sprintf("\nCurrent size: %v %v\n", width, height)
		s += fmt.Sprintf("\nPages: %v \n", m.openedFileText.Pages)
		s += "\n"
		s += "\nTo go back to the main menu, press 'b'. \n"
	}

	// Send the UI for rendering
	return s
}

func main() {
	err := filepath.WalkDir("./texts", visitFile)
	if err != nil {
		fmt.Print("All right")
	}
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
