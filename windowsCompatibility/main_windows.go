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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nsf/termbox-go"
)

// Styles for the app

var (
	// This is the style for the titles in the menus
	titleStyle = lipgloss.NewStyle().MarginLeft(2)
	// Style for the items to select in the menus
	itemStyle = lipgloss.NewStyle().PaddingLeft(4)
	// Style for the current selected item in the menu
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	// Style for the text to quit
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	// Styles for the different levels of word knowledge
	notKnownItemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")) // not known --> red
	semiKnownItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCA3A")) // semiknown --> yellow
	knownItemStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b300")) // known --> green
)

// visitFile function:
// prints out the files inside a directory

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

// listDirectories function:
// prints out the directories inside a directory

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
	cursor2         int    // language selection cursor
	currentLanguage string // current selected language
	currentError    string // current error we're encountering
	bootLanguage    string // the language we selected for the interface
	// counts the translation requests made so far: a result is shown only if it belongs to the latest request
	translateRequest int
	makingDictionary bool // true while a dictionary file is being created, so that it is not started twice
}

// Messages sent back to Update by the commands that run the slow operations (API calls, audio playback)
// in the background, so that the interface keeps responding while waiting for them.

type translationMsg struct {
	request     int
	translation string
	err         string
}

type audioMsg struct {
	err string
}

// translateCmd runs the translation in the background.
func translateCmd(request int, word string, languageId string, bootLanguage string) tea.Cmd {
	return func() tea.Msg {
		translation, errString := translator.Translate(word, languageId, bootLanguage)
		return translationMsg{request: request, translation: translation, err: errString}
	}
}

// audioCmd downloads (if not cached) and plays the audio of a word in the background.
func audioCmd(word string, languageId string) tea.Cmd {
	return func() tea.Msg {
		mp3FilePath, audioErr := audioPlayer.GetAudio(word, languageId)
		// the mp3 stays inside the audio cache, so it is not downloaded again next time
		if audioErr == "" {
			audioErr = audioPlayer.PlayMP3(mp3FilePath)
		}
		return audioMsg{err: audioErr}
	}
}

type dictionaryMsg struct {
	language string
	err      string
}

// dictionaryCmd creates the dictionary file in the background (one API call per word we don't know).
func dictionaryCmd(wordLevels map[string]int, language string, bootLanguage string) tea.Cmd {
	// copy the map, since Update keeps changing the original while the dictionary is being made
	data := make(map[string]int, len(wordLevels))
	for k, v := range wordLevels {
		data[k] = v
	}
	return func() tea.Msg {
		return dictionaryMsg{language: language, err: fileReader.MakeDictionary(data, language, bootLanguage)}
	}
}

// startDictionary starts dictionaryCmd, unless a dictionary is already being created.
func (m model) startDictionary(wordLevels map[string]int) (tea.Model, tea.Cmd) {
	if m.makingDictionary {
		return m, nil
	}
	m.makingDictionary = true
	m.currentError = fmt.Sprintf("languages/%s/dictionary.txt ...", m.currentLanguage)
	return m, dictionaryCmd(wordLevels, m.currentLanguage, m.bootLanguage)
}

// initialModel function:
// inits a model struct for the app

func initialModel() model {
	bootLang, _ := ioutil.ReadFile("setup/bootLanguage.txt")
	bootLangString := string(bootLang)
	directoryPath := "languages"

	directories, _ := listDirectories(directoryPath)
	directories = directories[1:]
	return model{
		// Our to-do list is a grocery list
		choices:      filePaths,
		choices2:     directories,
		viewIndex:    2,
		cursor2:      0,
		currentError: "",
		bootLanguage: bootLangString,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Results of the background commands
	switch msg := msg.(type) {
	case translationMsg:
		if msg.request == m.translateRequest {
			m.openedFileText.CurrentTranslate = msg.translation
			m.currentError = msg.err
		}
		return m, nil
	case audioMsg:
		m.currentError += msg.err
		return m, nil
	case dictionaryMsg:
		m.makingDictionary = false
		if msg.err != "" {
			m.currentError = msg.err
		} else {
			m.currentError = fmt.Sprintf("languages/%s/dictionary.txt ✓", msg.language)
		}
		return m, nil
	}

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
				m.openedFile = m.choices[m.cursor]
				text := fileReader.InitText(m.openedFile, m.currentLanguage)
				m.openedFileText = text
				// discard the translation still pending for the previously opened text
				m.translateRequest++
			case "f":
				return m.startDictionary(fileReader.MakeDictFromMenu(m.currentLanguage))
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
				m.openedFileText.SetCurrentWordLevel(0, m.currentLanguage)
			case "1":
				m.openedFileText.SetCurrentWordLevel(1, m.currentLanguage)
			case "2":
				m.openedFileText.SetCurrentWordLevel(2, m.currentLanguage)
			case "3":
				m.openedFileText.SetCurrentWordLevel(3, m.currentLanguage)

			case "4":
				currentLanguageId := languageHandler.LanguageMap[m.currentLanguage]
				m.currentError = ""
				return m, audioCmd(m.openedFileText.CurrentWord(), currentLanguageId)

			// get translation
			case "5":
				currentlLanguageId := languageHandler.LanguageMap2[m.currentLanguage]
				m.translateRequest++
				m.openedFileText.CurrentTranslate = "..."
				return m, translateCmd(m.translateRequest, m.openedFileText.CurrentWord(), currentlLanguageId, m.bootLanguage)

			case "f":
				return m.startDictionary(m.openedFileText.WordLevels)

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
				m.currentLanguage = m.choices2[m.cursor2][10:]

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
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][4]
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][5]
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][1] + "\n"
		if m.currentError != "" {
			s += m.currentError + "\n"
		}
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
			} else if value, ok := m.openedFileText.WordLevels[m.openedFileText.WordList[actualKey]]; ok {
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
		s = ""
		s += interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][0] + "\n\n"

		// Iterate over our choices

		for i, choice := range m.choices2 {

			// Is the cursor pointing at this choice?
			cursor := " " // no cursor
			if m.cursor2 == i {
				cursor = ">" // cursor!
			}

			// Is this choice selected?
			var checked = "x" // selected!

			// Render the row
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice[10:])
		}

		// The footer
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][1] + "\n"
	}

	// Send the UI for rendering
	return s
}

// main function

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	err := filepath.WalkDir("./texts", visitFile)
	if err != nil {
		fmt.Print("All right")
	}
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	termbox.Close()
}
