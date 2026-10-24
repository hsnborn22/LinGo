package main

// This is the main file: when you will run the compiled executable the code inside the main function
// is what will run.

// Imports:

/*
1) fmt --> for printing, formatting exc.
2) io/ioutil and os --> working with files
3) path/filepath --> used to list the number of subdirectories and files inside directories.
4) strings --> used for some string methods used throughout the program
5) all the packages with example.com in front are the packages that can be found in the respective folders:
the files inside will explain what each package does.

6) Bubbletea and lipgloss --> used for the interface.
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"example.com/packages/audioPlayer"
	"example.com/packages/fileReader"
	"example.com/packages/interfaceLanguage"
	"example.com/packages/languageHandler"
	"example.com/packages/strokeOrder"
	"example.com/packages/terminalSize"
	"example.com/packages/translator"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
)

// Styles for the app
// The styles were done using lipgloss; the titles are pretty self-explanatory but I will explain what each of these
// styles do.

var (
	titleStyle = lipgloss.NewStyle().MarginLeft(2)
	itemStyle  = lipgloss.NewStyle().PaddingLeft(4)
	// Style for the current item we're hovering over in a text.
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	// This is the style for the "quit text", i.e the text that tells us how to quit the program.
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	// These are the styles for all the "word levels"
	// 1) Not known --> word will become red
	notKnownItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	// 2) Semi-known --> word will become yellow
	semiKnownItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCA3A"))
	// 3) Known --> word will become green
	knownItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b300"))
)

// These are the styles for the tables used in the menus

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// visitFile function:
// this function appends the list of files inside a given directory inside the filePaths slice.
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
// this function lists the subdirectories inside a certain path. it returns the list in the
// form of a slice and a possible error (which is nil if everything went as expected)

func listDirectories(directoryPath string) ([]string, error) {
	// declare the slice variable we're going to return
	var directories []string

	// here we're using the filepath package we imported at the beginning
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// if the files we're encountering while "walking" are a directories, append them to the directories slice.
			directories = append(directories, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// return the slice and the error (which if we reached this point will be nil since no error occured)

	return directories, nil
}

var filePaths []string // Declare a global slice to store file paths

type model struct {
	choices         []string        // text-file selection menu (OLD)
	choices2        []string        // language select menu (OLD)
	cursor          int             // which to-do list item our cursor is pointing at
	viewIndex       int             // viewIndex --> will be 0 for the menu, and 1 for an opened file.
	openedFile      string          // will store the name of the file we opened.
	openedFileText  fileReader.Text // will store the fileReader.Text object representing the file we opened
	cursor2         int             //
	currentLanguage string          // This is the current language we're studying
	currentError    string          // This is the current error that go cathced; in this way, if something goes wrong with the APIs the app
	// doesn't crash, it just notifies the user and tells you what went wrong.
	bootLanguage  string              // This is the language of the interface (if you're reading this I assume you speak english and thus it will be english)
	languageTable table.Model         // This is the table listing all the languages (new UI)
	textTable     table.Model         // This is the table listing all the text files we can open inside the app (new UI).
	hanziData     map[string][]string // This is the map that stores the pinyin equivalent of the most common hanzi in simplified mandarin chinese
	// These count the translation/latinization requests made so far: a result is shown only if it belongs to
	// the latest request, so that a slow answer for an old word never overwrites the one for the current word.
	translateRequest int
	latinizeRequest  int
	makingDictionary bool   // true while a dictionary file is being created, so that it is not started twice
	defaultLanguage  string // The language selected as default: when set, the language selection menu is skipped at startup
	defaultCursor    int    // Cursor of the menu shown after selecting a language (0 --> continue, 1 --> set/remove as default)
}

// File storing the default language (it doesn't exist if no default language was chosen).
const defaultLanguageFile = "setup/defaultLanguage.txt"

// Messages sent back to Update by the commands that run the slow operations (API calls, audio playback)
// in the background, so that the interface keeps responding while waiting for them.

type translationMsg struct {
	request     int
	translation string
	err         string
}

type latinizationMsg struct {
	request      int
	latinization string
}

type audioMsg struct {
	err string
}

type dictionaryMsg struct {
	language string
	err      string
}

// translateCmd runs the translation in the background; useGoogle selects Translate2 instead of Translate.
func translateCmd(request int, word string, languageId string, bootLanguage string, useGoogle bool) tea.Cmd {
	return func() tea.Msg {
		var translation, errString string
		if useGoogle {
			translation, errString = translator.Translate2(word, languageId, bootLanguage)
		} else {
			translation, errString = translator.Translate(word, languageId, bootLanguage)
		}
		return translationMsg{request: request, translation: translation, err: errString}
	}
}

// latinizeCmd runs the latinization in the background (some languages need an API call).
func latinizeCmd(request int, word string, hanziData map[string][]string, language string) tea.Cmd {
	return func() tea.Msg {
		return latinizationMsg{request: request, latinization: translator.LatinizeText(word, hanziData, language)}
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

// dictionaryCmd creates the dictionary file in the background (one API call per word we don't know);
// withLatinization selects MakeAltDictionary instead of MakeDictionary.
func dictionaryCmd(wordLevels map[string]int, language string, bootLanguage string, hanziData map[string][]string, withLatinization bool) tea.Cmd {
	// copy the map, since Update keeps changing the original while the dictionary is being made
	data := make(map[string]int, len(wordLevels))
	for k, v := range wordLevels {
		data[k] = v
	}
	return func() tea.Msg {
		if withLatinization {
			return dictionaryMsg{language: language, err: fileReader.MakeAltDictionary(data, language, bootLanguage, hanziData)}
		}
		return dictionaryMsg{language: language, err: fileReader.MakeDictionary(data, language, bootLanguage)}
	}
}

// startDictionary starts dictionaryCmd, unless a dictionary is already being created.
func (m model) startDictionary(wordLevels map[string]int, withLatinization bool) (tea.Model, tea.Cmd) {
	if m.makingDictionary {
		return m, nil
	}
	m.makingDictionary = true
	m.currentError = fmt.Sprintf("languages/%s/dictionary.txt ...", m.currentLanguage)
	return m, dictionaryCmd(wordLevels, m.currentLanguage, m.bootLanguage, m.hanziData, withLatinization)
}

// This function initializes the bubbletea model to boot the application; this is one of the "dirtiest" parts of the application
// in terms of code and it needs a serious refactor.

func initialModel() model {
	// Get the boot language from the bootLanguage.txt file by reading it.
	bootLang, _ := ioutil.ReadFile("setup/bootLanguage.txt")
	// Convert the []byte object we got into a string
	bootLangString := string(bootLang)
	// These 4 lines of code just make sure to remove all the possible spaces, new lines and tabular spaces.
	stripped := strings.ReplaceAll(bootLangString, "\t", "")
	stripped = strings.ReplaceAll(stripped, " ", "")
	stripped = strings.ReplaceAll(stripped, "\n", "")
	bootLangString = stripped

	languageTitle := interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[bootLangString]][0]
	columns := []table.Column{
		{Title: languageTitle, Width: 35},
	}
	directoryPath := "languages"

	// Store the subdirectories of the "languages" directory inside the directories slice;
	// this lists all the languages we can study with this app.
	directories, _ := listDirectories(directoryPath)
	// skip the first one tho because it will be the directory "languages" itself;
	// (this is because of how the filepath package works)
	directories = directories[1:]

	// All this code just creates the tables for the UI; I will place this into another file soon and refactor it.
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

	// Read the default language, if one was chosen; if it's still one of the available languages,
	// start directly in the text selection menu and point the language table to it.
	defaultLang, _ := ioutil.ReadFile(defaultLanguageFile)
	defaultLangString := strings.TrimSpace(string(defaultLang))
	startView := 2
	startLanguage := ""
	for i, row := range rows {
		if defaultLangString != "" && row[0] == defaultLangString {
			startView = 0
			startLanguage = defaultLangString
			t.SetCursor(i)
		}
	}
	if startLanguage == "" {
		defaultLangString = ""
	}

	textTitle := interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[bootLangString]][3]
	columns2 := []table.Column{
		{Title: textTitle, Width: 35},
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

	// return the model object we need to start the bubbletea app.
	return model{
		choices:         filePaths,
		choices2:        directories,
		viewIndex:       startView,
		cursor2:         0,
		currentError:    "",
		currentLanguage: startLanguage,
		defaultLanguage: defaultLangString,
		bootLanguage:    bootLangString,
		languageTable:   t,
		textTable:       t2,
		hanziData:       translator.InitHanzi(),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// Update method; this update method handles the changes that need to be done to the UI
// when a key is pressed or when something is changed (in according with the ELM architecture principles).

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.openedFileText.PageList = fileReader.DivideInPages(m.openedFileText.TokenList)
	m.openedFileText.Pages = len(m.openedFileText.PageList)
	if m.openedFileText.CurrentPage > m.openedFileText.Pages-1 {
		m.openedFileText.CurrentPage = 0
	}
	// Only the table of the menu we're in reacts to the keys
	switch m.viewIndex {
	case 0:
		m.textTable, _ = m.textTable.Update(msg)
	case 2:
		m.languageTable, _ = m.languageTable.Update(msg)
	}

	// Results of the background commands
	switch msg := msg.(type) {
	case translationMsg:
		if msg.request == m.translateRequest {
			m.openedFileText.CurrentTranslate = msg.translation
			m.currentError = msg.err
		}
		return m, nil
	case latinizationMsg:
		if msg.request == m.latinizeRequest {
			m.openedFileText.CurrentLatinization = msg.latinization
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
				m.openedFile = "texts/" + m.textTable.SelectedRow()[0]
				text := fileReader.InitText(m.openedFile, m.currentLanguage)
				m.openedFileText = text
				// discard the results still pending for the previously opened text
				m.translateRequest++
				m.latinizeRequest++
			// If the key pressed is f, generate a dictionary file.
			case "f":
				return m.startDictionary(fileReader.MakeDictFromMenu(m.currentLanguage), false)
			case "z":
				return m.startDictionary(fileReader.MakeDictFromMenu(m.currentLanguage), true)
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

			// We use a and d to move between pages
			// If the key pressed is d, move one page to the right
			case "d":
				// Check that we're not on the last page first; if we were on the last page
				// we would ideally not want to be able to go one page further since there's no
				// next page by definition
				//fmt.Println(m.openedFileText.CurrentPage, m.openedFileText.Pages)
				if m.openedFileText.CurrentPage < m.openedFileText.Pages-1 {
					// If everything is alright implement the logic, i.e augment the page counter and go to the next page
					m.openedFileText.CurrentPage++
				}
			// If the key pressed is a, move one page to the left
			case "a":
				// Check if we're not on page 0 first: if we were, we would go to page -1 and that doesn't make sense.
				if m.openedFileText.CurrentPage > 0 {
					// If everything is alright implement the logic, i.e decrement the page counter and go to the next page
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
				return m, translateCmd(m.translateRequest, m.openedFileText.CurrentWord(), currentlLanguageId, m.bootLanguage, true)

			case "6":
				m.latinizeRequest++
				m.openedFileText.CurrentLatinization = "..."
				return m, latinizeCmd(m.latinizeRequest, m.openedFileText.CurrentWord(), m.hanziData, m.currentLanguage)
			case "7":
				url := fmt.Sprintf("https://www.strokeorder.com/chinese/%s", m.openedFileText.CurrentWord())
				err := strokeOrder.OpenBrowser(url)
				if err != nil {
					m.currentError += err.Error()
				}
			case "8":
				url := fmt.Sprintf("https://translate.google.com/?sl=%s&tl=%s&text=%s&op=translate", languageHandler.LanguageMap2[m.currentLanguage], m.bootLanguage, m.openedFileText.CurrentWord())
				err := strokeOrder.OpenBrowser(url)
				if err != nil {
					m.currentError += err.Error()
				}
			case "9":
				currentlLanguageId := languageHandler.LanguageMap2[m.currentLanguage]
				m.translateRequest++
				m.openedFileText.CurrentTranslate = "..."
				return m, translateCmd(m.translateRequest, m.openedFileText.CurrentWord(), currentlLanguageId, m.bootLanguage, false)

			case "f":
				return m.startDictionary(m.openedFileText.WordLevels, false)
			case "z":
				return m.startDictionary(m.openedFileText.WordLevels, true)

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
				m.viewIndex = 3
				m.defaultCursor = 0
				m.currentLanguage = m.languageTable.SelectedRow()[0]

			}

		}
	// Menu shown after selecting a language: continue, or set (remove) it as the default language.
	case 3:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.defaultCursor > 0 {
					m.defaultCursor--
				}
			case "down", "j":
				if m.defaultCursor < 1 {
					m.defaultCursor++
				}
			case "b":
				m.viewIndex = 2
				m.currentLanguage = ""
			case "enter", " ":
				m.viewIndex = 0
				m.currentError = ""
				if m.defaultCursor == 1 {
					if m.defaultLanguage == m.currentLanguage {
						// The language is already the default one: remove it as default.
						if err := os.Remove(defaultLanguageFile); err != nil && !os.IsNotExist(err) {
							m.currentError = err.Error()
						} else {
							m.defaultLanguage = ""
						}
					} else if err := os.WriteFile(defaultLanguageFile, []byte(m.currentLanguage+"\n"), 0644); err != nil {
						m.currentError = err.Error()
					} else {
						m.defaultLanguage = m.currentLanguage
					}
				}
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
		s += "\n"

		// Iterate over our choices

		s += baseStyle.Render(m.textTable.View())

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
		s += fmt.Sprintf("%v", m.openedFileText.CurrentPage)
		s += fmt.Sprintf("\n%s %v", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][9], m.openedFileText.Pages)
		s += "\n"
		s += fmt.Sprintf("%s %s", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][10], m.openedFileText.CurrentTranslate)
		s += "\n"
		s += fmt.Sprintf("%s %s", interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][13], m.openedFileText.CurrentLatinization)
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][11] + m.currentError
		s += "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][12] + "\n" + interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]][1]
	} else if m.viewIndex == 2 {
		return baseStyle.Render(m.languageTable.View()) + "\n"
	} else if m.viewIndex == 3 {
		interfaceText := interfaceLanguage.InterfaceLanguage[interfaceLanguage.LanguagesCodeMap[m.bootLanguage]]
		options := []string{interfaceText[15], interfaceText[16]}
		if m.defaultLanguage == m.currentLanguage {
			options[1] = interfaceText[17]
		}
		s = interfaceText[14] + m.currentLanguage + "\n\n"
		for i, option := range options {
			if i == m.defaultCursor {
				s += selectedItemStyle.Render("> "+option) + "\n"
			} else {
				s += itemStyle.Render(option) + "\n"
			}
		}
		s += "\n" + interfaceText[5] + "\n" + interfaceText[1] + "\n"
	}

	// Send the UI for rendering
	return s
}

// This is the main function, the main entrypoint of the program.

func main() {
	// Run the WalkDir function we looked at the beginning of the file inside of the texts directory
	// store the possible error we could get inside of the err variable
	err := filepath.WalkDir("./texts", visitFile)
	// If there is no error, it's all right; print it out to the console.
	if err != nil {
		fmt.Print("All right")
	}
	// Create a new bubbleTea program using the model returned by the initalModel() function
	// We saw what that function does at the beginning of the program.
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	// If there's an error in the running of the application, let the user know
	// by printing it out to the console.
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
