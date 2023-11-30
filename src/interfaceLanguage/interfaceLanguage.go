package interfaceLanguage

// InterfaceLanguage slice:
// Stores the translated interace in various languages.

var InterfaceLanguage [][]string = [][]string{
	{"What language do you want to study?", "Press q to quit.", "You are currently studying: ", "What text file do you want to open?\n\n", "\nPress f to make a dictionary file.\n", "Press b to go back to the language selection menu.\n", "Hello you are in ", " and cursor is at: ", "Current size: ", "Pages: ", "Translation of the selected word: ", "Error flag: ", "To go back to the main menu, press 'b' || Press f to make a dictionary file."},
	{"Che lingua vuoi studiare?", "Premere q per uscire.", "Stai studiando: ", "Che file di testo vuoi aprire?\n\n", "\nPremere 'f' per creare un file da esportare in flashcards.\n", "Premere 'b' per tornare al menu di selezione lingua.\n", "Sei correntemente in ", " e il cursore è alla posizione: ", "Dimensione attuale: ", "Pagine: ", "Traduzione della parola selezionata: ", "Errori: ", "Per tornare al menu principale, premere 'b' || Premere f per creare un flashcard file."},
	{"Quelle langue voulez-vous étudier?", "Appuyez sur la touche 'q' pour quitter.", "Vous étudiez maintenant: ", "Quel text-file voules-vouz ouvrir?\n\n", "\nAppuyez sur la touche 'f' pour créer une file pour le flashcards.\n", "Appuyez sur la touche 'b' pour retourner à le menu pour la selection d'une langue", "Vous êtes maintenant dans ", " et le curseur est à la position: ", "Dimension actuelle: ", "Pages: ", "Traduction de le mot sélectionné: ", "Erreurs: ", "Pour tourner à le menu principal, appuyez sur la touche 'b' || Appuyez sur la touche 'f' pour créer une file pour le flashcards."},
	{"¿Qué idioma quieres estudiar?", "Pulse 'q' para salir del programa.", "Actualmente estás estudiando: ", "¿Qué text-file quieres abrir?\n\n", "\nPulsa 'f' para crear un file para flashcards.\n", "Prensa 'b' para volver al menú de selección de idioma.", "Actualmente te encuentras en ", " y el cursor está en la posición: ", "Dimensiones actuales: ", "Paginas: ", "Traducción de la palabra seleccionada: ", "Errores: ", "Para volver a le menu principal, pulse 'b' || Pulsa 'f' para crear un file para flashcards. "},
	{"Welche Sprache möchtest du lernen?",
		"Drücke 'q', um das Programm zu beenden.",
		"Du lernst derzeit: ",
		"Welche Textdatei möchtest du öffnen?\n\n",
		"\nDrücke 'f', um eine Datei für Lernkarten zu erstellen.\n",
		"Drücke 'b', um zum Sprachauswahlmenü zurückzukehren.",
		"Du befindest dich derzeit in ",
		" und der Cursor befindet sich an der Position: ",
		"Aktuelle Abmessungen: ",
		"Seiten: ",
		"Übersetzung des ausgewählten Worts: ",
		"Fehler: ",
		"Um zum Hauptmenü zurückzukehren, drücke 'b' || Drücke 'f', um eine Datei für Lernkarten zu erstellen."},
	{"Какой язык вы хотите изучать?",
		"Нажмите 'q', чтобы выйти из программы.",
		"Сейчас изучаете: ",
		"Какой текстовый файл хотите вы открыть?\n\n",
		"\nНажимать 'f', чтобы создать файл для карточек.",
		"Нажимать 'b', чтобы вернуться в меню выбора языка.",
		"В настоящее время вы находитесь в ",
		" и курсор находится на позиции: ",
		"Размеры терминала: ",
		"Страницы: ",
		"Перевод выбранного слова: ",
		"Ошибки и исключения: ",
		"Для возврата в главное меню нажмите 'b' || Нажмите 'f', чтобы создать файл для флашкарды."},
}

// LanguagesCodeMap map:
// Stores the corresponding index for the translation inside the InterfaceLanguage slice
// of the interface in a particular language

var LanguagesCodeMap map[string]int = map[string]int{
	"en": 0,
	"it": 1,
	"fr": 2,
	"es": 3,
	"de": 4,
	"ru": 5,
}
