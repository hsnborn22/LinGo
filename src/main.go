package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"example.com/packages/audioPlayer"
	"example.com/packages/fileReader"
	"example.com/packages/interfaceLanguage"
	"example.com/packages/languageHandler"
	"example.com/packages/terminalSize"
	"example.com/packages/translator"
	"github.com/charmbracelet/bubbles/table"
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

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

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

func listDirectories(directoryPath string) ([]string, error) {
	var directories []string

	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return directories, nil
}

var filePaths []string // Declare a global slice to store file paths

type model struct {
	choices         []string // items on the to-do list
	choices2        []string // language select menu
	cursor          int      // which to-do list item our cursor is pointing at
	viewIndex       int      // viewIndex --> will be 0 for the menu, and 1 for an opened file.
	openedFile      string   // will store the name of the file we opened.
	openedFileText  fileReader.Text
	cursor2         int
	currentLanguage string
	currentError    string
	bootLanguage    string
	languageTable   table.Model
	textTable       table.Model
}

func initialModel() model {
	columns := []table.Column{
		{Title: "Select a language", Width: 20},
	}
	directoryPath := "languages"

	directories, _ := listDirectories(directoryPath)
	directories = directories[1:]
	var rows []table.Row

	for _, v := range directories {
		e := table.Row{v[10:]}
		rows = append(rows, e)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	columns2 := []table.Column{
		{Title: "Select a text", Width: 20},
	}

	var rows2 []table.Row

	for _, v := range filePaths {
		e := table.Row{v[6:]}
		rows2 = append(rows2, e)
	}

	t2 := table.New(
		table.WithColumns(columns2),
		table.WithRows(rows2),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	t2.SetStyles(s)
	bootLang, _ := ioutil.ReadFile("setup/bootLanguage.txt")
	bootLangString := string(bootLang)
	return model{
		// Our to-do list is a grocery list
		choices:       filePaths,
		choices2:      directories,
		viewIndex:     2,
		cursor2:       0,
		currentError:  "",
		bootLanguage:  bootLangString,
		languageTable: t,
		textTable:     t2,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.languageTable, _ = m.languageTable.Update(msg)
	m.textTable, _ = m.textTable.Update(msg)
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
			case "b":
				m.viewIndex = 2
				m.currentLanguage = ""

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "enter", " ":
				m.viewIndex = 1
				m.openedFile = "texts/" + m.textTable.SelectedRow()[0]
				text := fileReader.InitText(m.openedFile, m.currentLanguage)
				m.openedFileText = text
			case "f":
				dictionary := fileReader.MakeDictFromMenu(m.currentLanguage)
				fileReader.MakeDictionary(dictionary, m.currentLanguage, m.bootLanguage)
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

			case "0":
				m.openedFileText.WordLevels[m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition]] = 0
				fileReader.MakeJsonFile(m.openedFileText.WordLevels, m.currentLanguage)
			case "1":
				m.openedFileText.WordLevels[m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition]] = 1
				fileReader.MakeJsonFile(m.openedFileText.WordLevels, m.currentLanguage)
			case "2":
				m.openedFileText.WordLevels[m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition]] = 2
				fileReader.MakeJsonFile(m.openedFileText.WordLevels, m.currentLanguage)
			case "3":
				m.openedFileText.WordLevels[m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition]] = 3
				fileReader.MakeJsonFile(m.openedFileText.WordLevels, m.currentLanguage)

			case "4":
				currentLanguageId := languageHandler.LanguageMap[m.currentLanguage]
				m.currentError = ""
				m.currentError += audioPlayer.GetAudio(m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition], currentLanguageId)
				mp3FilePath := fmt.Sprintf("audio/%s.mp3", m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition])

				if m.currentError != "" {
					m.currentError += "\n"
				}
				m.currentError += audioPlayer.PlayMP3(mp3FilePath)
				m.currentError += "\n"
				m.currentError += audioPlayer.DeleteMP3(mp3FilePath)

			// get translation
			case "5":
				currentlLanguageId := languageHandler.LanguageMap2[m.currentLanguage]
				translation, errString := translator.Translate(m.openedFileText.TokenList[m.openedFileText.TokenCursorPosition], currentlLanguageId, m.bootLanguage)
				m.currentError = errString
				m.openedFileText.CurrentTranslate = translation

			case "f":
				fileReader.MakeDictionary(m.openedFileText.WordLevels, m.currentLanguage, m.bootLanguage)

			// Move the cursor to the beginning of the current page.
			case "m":
				currentCursor := m.openedFileText.CurrentPage * terminalSize.GetLinesPerPage() * terminalSize.GetWordsPerLine()
				m.openedFileText.TokenCursorPosition = currentCursor

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "b":
				m.viewIndex = 0
				m.currentError = ""
			}
		}
	case 2:
		switch msg := msg.(type) {

		// Is it a key press?
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor2 > 0 {
					m.cursor2--
				}

			// The "down" and "j" keys move the cursor2 down
			case "down", "j":
				if m.cursor2 < len(m.choices2)-1 {
					m.cursor2++
				}

			// The "enter" key and the spacebar (a literal space) toggle
			// the selected state for the item that the cursor is pointing at.
			case "enter", " ":
				m.viewIndex = 0
				m.currentLanguage = m.languageTable.SelectedRow()[0]

			}

		}

	}
	return m, nil
}

func (m model) View() string {
	var s string
	if m.viewIndex == 0 {
		// The header
		s = interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][2] + m.currentLanguage + "\n"
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][3]

		// Iterate over our choices

		s += baseStyle.Render(m.textTable.View())

		// The footer
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][4]
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][5]
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][1] + "\n"
	} else if m.viewIndex == 1 {
		wordsPerLine := terminalSize.GetWordsPerLine()
		linesPerPage := terminalSize.GetLinesPerPage()
		width, height := terminalSize.GetTerminalSize()

		s = interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][6] + m.openedFile + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][7]
		s += fmt.Sprintf("%v", m.openedFileText.TokenCursorPosition)
		s += fmt.Sprintf("\n%s %v %v\n", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][8], width, height)
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
		s += fmt.Sprintf("\n%s %v", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][9], m.openedFileText.Pages)
		s += "\n"
		s += fmt.Sprintf("%s %s", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][10], m.openedFileText.CurrentTranslate)
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][11] + m.currentError
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][12] + "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][1]
	} else if m.viewIndex == 2 {
		return baseStyle.Render(m.languageTable.View()) + "\n"
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
