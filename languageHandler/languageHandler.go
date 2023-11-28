/*
	=====================================================================

** languageHandler package **
This package is responsible for the mapping of determinate languages to their
IDs, which can then be used in the APIs employed by the application.

    =====================================================================
*/

package languageHandler

// the LanguageMap map contains the languages mapped to their IDs.

var LanguageMap map[string]string = map[string]string{
	"afrikaans":   "af-ZA",
	"albanian":    "sq",
	"arabic":      "ar-AE",
	"armenian":    "hy",
	"bengali-bd":  "bn-BD",
	"bengali-in":  "bn-IN",
	"bosnian":     "bs",
	"burmese":     "my",
	"catalan":     "ca-ES",
	"chinese":     "cmn-Hant-TW",
	"croatian":    "hr-HR",
	"czech":       "cs-CZ",
	"danish":      "da-DK",
	"dutch":       "nl-NL",
	"english-aus": "en-AU",
	"english-gb":  "en-GB",
	"english-us":  "en-US",
	"esperanto":   "eo",
	"estonian":    "et",
	"filipino":    "fil-PH",
	"finnish":     "fi-FI",
	"french":      "fr-FR",
	"french-can":  "fr-CA",
	"german":      "de-DE",
	"greek":       "el-GR",
	"gujarati":    "gu",
	"hindi":       "hi-IN",
	"hungarian":   "hu-HU",
	"icelandic":   "is-IS",
	"indonesian":  "id-ID",
	"italian":     "it-IT",
	"japanese":    "ja-JP",
	"kannada":     "kn",
	"khmer":       "km",
	"korean":      "ko-KR",
	"latin":       "la",
	"latvian":     "lv",
	"macedonian":  "mk",
	"marathi":     "mr",
	"malayalam":   "ml",
	"nepali":      "ne",
	"norwegian":   "nb-NO",
	"polish":      "pl-PL",
	"portuguese":  "pt-BR",
	"romanian":    "ro-RO",
	"russian":     "ru-RU",
	"serbian":     "sr-RS",
	"slovak":      "sk-SK",
	"spanish":     "es-ES",
	"swedish":     "sv-SE",
	"turkish":     "tr-TR",
	"ukrainian":   "uk-UA",
	"vietnamese":  "vi-VN",
	"welsh":       "cy",
}
