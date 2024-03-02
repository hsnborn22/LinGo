package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

// This is the map that contains how cyrillic letters in all variations of cyrillic
// (russian,mongolian,ukrainian, serbian,kazakh,tajik exc.) are mapped to latin.

var cyrillicToLatin map[string]string = map[string]string{"А": "A", "Б": "B", "В": "V", "Г": "G", "Д": "D", "Ђ": "Đ", "Е": "E", "Ё": "YO", "Ж": "ZH", "З": "Z", "И": "I", "Й": "Y", "К": "K", "Л": "L", "Љ": "LJ", "М": "M", "Н": "N", "Њ": "NJ", "О": "O", "П": "P", "Р": "R", "С": "S", "Т": "T", "У": "U", "Ф": "F", "Х": "KH", "Ц": "TS", "Ч": "CH", "Ш": "SH", "Щ": "SHCH", "Ъ": "''", "Ы": "Y", "Ь": "'", "Э": "E", "Ю": "YU", "Я": "YA",
	"а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "ђ": "đ", "е": "e", "ё": "yo", "ж": "zh", "з": "z", "и": "i", "й": "y", "к": "k", "л": "l", "љ": "lj", "м": "m", "н": "n", "њ": "nj", "о": "o", "п": "p", "р": "r", "с": "s", "т": "t", "у": "u", "ф": "f", "х": "kh", "ц": "ts", "ч": "ch", "ш": "sh", "щ": "shch", "ъ": "''", "ы": "y", "ь": "'", "э": "e", "ю": "yu", "я": "ya",
	"Ң": "NG", "Ү": "U", "Ұ": "U", "Һ": "H", "Ө": "O", "ү": "u", "ұ": "u", "һ": "h", "ө": "o",
	"Ә": "A", "Ғ": "G", "Қ": "Q", "ә": "a", "ғ": "g", "қ": "q", "ң": "n",
}

// This is the map that contains how to convert georgian script to latin
var georgianToLatin = map[rune]string{
	'ა': "a", 'ბ': "b", 'გ': "g", 'დ': "d", 'ე': "e",
	'ვ': "v", 'ზ': "z", 'თ': "t", 'ი': "i", 'კ': "k'",
	'ლ': "l", 'მ': "m", 'ნ': "n", 'ო': "o", 'პ': "p'",
	'ჟ': "zh", 'რ': "r", 'ს': "s", 'ტ': "t'", 'უ': "u",
	'ფ': "p", 'ქ': "k", 'ღ': "gh", 'ყ': "q'", 'შ': "sh",
	'ჩ': "ch", 'ც': "ts", 'ძ': "dz", 'წ': "ts'", 'ჭ': "ch'",
	'ხ': "kh", 'ჯ': "j", 'ჰ': "h",
}

// Define the mapping of Hindi characters to Latin characters
var HindiReplacements = map[string]string{
	"क": "k", "ख": "kh", "ग": "ga", "घ": "gh", "ङ": "ng",
	"च": "ch", "छ": "chh", "ज": "j", "झ": "jh", "ञ": "ny",
	"ट": "t", "ठ": "th", "ड": "d", "ढ": "dh", "ण": "n",
	"त": "t", "थ": "th", "द": "d", "ध": "dh", "न": "n",
	"प": "p", "फ": "f", "ब": "b", "भ": "bh", "म": "m",
	"य": "y", "र": "r", "ल": "l", "व": "v", "श": "sh",
	"ष": "s", "स": "s", "ह": "h", "क़": "k", "ख़": "kh",
	"ग़": "g", "ऩ": "n", "ड़": "d", "ढ़": "rh",
	"ऱ": "r", "य़": "ye", "ळ": "l", "ऴ": "ll", "फ़": "f",
	"ज़": "z", "ऋ": "ri", "ा": "aa", "ि": "i", "ी": "i",
	"ु": "u", "ू": "u", "ॅ": "e", "ॆ": "e", "े": "e",
	"ै": "ai", "ॉ": "o", "ॊ": "o", "ो": "o", "ौ": "au",
	"अ": "a", "आ": "aa", "इ": "i", "ई": "ee", "उ": "u",
	"ऊ": "oo", "ए": "e", "ऐ": "ai", "ऑ": "au", "ओ": "o",
	"औ": "au", "ँ": "n", "ं": "n", "ः": "ah", "़": "e",
	"्": "", "०": "0", "१": "1", "२": "2", "३": "3",
	"४": "4", "५": "5", "६": "6", "७": "7", "८": "8",
	"९": "9", "।": ".", "ऍ": "e", "ृ": "ri", "ॄ": "rr",
	"ॠ": "r", "ऌ": "l", "ॣ": "l", "ॢ": "l", "ॡ": "l",
	"ॿ": "b", "ॾ": "d", "ॽ": "", "ॼ": "j", "ॻ": "g",
	"ॐ": "om", "ऽ": "'", "e.a": "a", "\n": "\n",
}

// This is the map that contains how greek characters are mapped phonetically to
// latin characters.

var greekToLatin = map[string]string{
	"Α": "A",
	"Β": "B",
	"Γ": "G",
	"Δ": "D",
	"Ε": "E",
	"Ζ": "Z",
	"Η": "H",
	"Θ": "TH",
	"Ι": "I",
	"Κ": "K",
	"Λ": "L",
	"Μ": "M",
	"Ν": "N",
	"Ξ": "X",
	"Ο": "O",
	"Π": "P",
	"Ρ": "R",
	"Σ": "S",
	"Τ": "T",
	"Υ": "U",
	"Φ": "PH",
	"Χ": "CH",
	"Ψ": "PS",
	"Ω": "O",
	"α": "a",
	"β": "b",
	"γ": "g",
	"δ": "d",
	"ε": "e",
	"ζ": "z",
	"η": "h",
	"θ": "th",
	"ι": "i",
	"κ": "k",
	"λ": "l",
	"μ": "m",
	"ν": "n",
	"ξ": "x",
	"ο": "o",
	"π": "p",
	"ρ": "r",
	"σ": "s",
	"τ": "t",
	"υ": "u",
	"φ": "ph",
	"χ": "ch",
	"ψ": "ps",
	"ω": "o",
}

// Mapping of Arabic to Latin characters
var arabicToLatin = map[string]string{
	"ا":  "a",
	"أ":  "a",
	"آ":  "a",
	"إ":  "e",
	"ب":  "b",
	"ت":  "t",
	"ث":  "th",
	"ج":  "j",
	"ح":  "h",
	"خ":  "kh",
	"د":  "d",
	"ذ":  "d",
	"ر":  "r",
	"ز":  "z",
	"س":  "s",
	"ش":  "sh",
	"ص":  "s",
	"ض":  "d",
	"ط":  "t",
	"ظ":  "z",
	"ع":  "'e",
	"غ":  "gh",
	"ف":  "f",
	"ق":  "q",
	"ك":  "k",
	"ل":  "l",
	"م":  "m",
	"ن":  "n",
	"ه":  "h",
	"و":  "w",
	"ي":  "y",
	"ى":  "a",
	"ئ":  "'e",
	"ء":  "'",
	"ؤ":  "'e",
	"لا": "la",
	"ة":  "h",
	"؟":  "?",
	"!":  "!",
	"ـ":  "",
	"،":  ",",
	"َ":  "a",
	"ُ":  "u",
	"ِ":  "e",
	"ٌ":  "un",
	"ً":  "an",
	"ٍ":  "en",
	"ّ":  "",
}

// Define the mapping of Farsi characters to Latin characters
var FarsiReplacements = map[string]string{
	"ا": "a", "أ": "a", "آ": "a", "إ": "e", "ب": "b",
	"ت": "t", "ث": "th", "ج": "j", "ح": "h", "خ": "kh",
	"د": "d", "ذ": "d", "ر": "r", "ز": "z", "س": "s",
	"ش": "sh", "ص": "s", "ض": "d", "ط": "t", "ظ": "z",
	"ع": "'e", "غ": "gh", "ف": "f", "ق": "q", "ك": "k",
	"ل": "l", "م": "m", "ن": "n", "ه": "h", "و": "w",
	"ي": "y", "ى": "a", "ئ": "'e", "ء": "'", "ؤ": "'e",
	"لا": "la", "ک": "ke", "پ": "pe", "چ": "che", "ژ": "je",
	"گ": "gu", "ی": "a", "ٔ": "", "ة": "h", "؟": "?",
	"!": "!", "ـ": "", "،": ",", "َ‎": "a", "ُ": "u",
	"ِ‎": "e", "ٌ": "un", "ً": "an", "ٍ": "en", "ّ": "",
	"\n": "\n",
}

// Basic mapping from Burmese characters to Latin alphabet.
var burmeseToLatinMap = map[string]string{
	// Consonants
	"က": "k", "ခ": "kh", "ဂ": "g", "ဃ": "gh", "င": "ng",
	"စ": "c", "ဆ": "ch", "ဇ": "z", "ဈ": "j", "ဉ": "ny",
	"ည": "ny", "ဋ": "t", "ဌ": "th", "ဍ": "d", "ဎ": "dh",
	"ဏ": "n", "တ": "t", "ထ": "th", "ဒ": "d", "ဓ": "dh",
	"န": "n", "ပ": "p", "ဖ": "ph", "ဗ": "b", "ဘ": "bh",
	"မ": "m", "ယ": "y", "ရ": "r", "လ": "l", "ဝ": "w",
	"သ": "s", "ဟ": "h", "ဠ": "l", "အ": "a",

	// Independent vowels
	"ဣ": "i", "ဤ": "ī", "ဥ": "u", "ဦ": "ū", "ဧ": "e",
	"ဩ": "o", "ဪ": "au",

	// Dependent vowel signs
	"ာ": "ā", "ိ": "i", "ီ": "ī", "ု": "u", "ူ": "ū",
	"ေ": "e", "ဲ": "ai", "ံ": "an", "့": "", "း": "",
	"္": "", // Used for stacking consonants
	"်": "", // Kill the inherent vowel of a consonant

	// Various diacritics
	"ျ": "ya", "ြ": "ra", "ွ": "wa", "ှ": "ha",
	"ဿ": "sa", "၀": "la",

	// Other symbols
	"၊": ",", "။": ".",

	// Numbers
	"၁": "1", "၂": "2", "၃": "3", "၄": "4",
	"၅": "5", "၆": "6", "၇": "7", "၈": "8", "၉": "9",
}

// Map that contains info about how to latinize armenian texts
var armenianToLatin = map[rune]string{
	'Ա': "A", 'ա': "a",
	'Բ': "B", 'բ': "b",
	'և': "ev",
	'Գ': "G", 'գ': "g",
	'Դ': "D", 'դ': "d",
	'Ե': "E", 'ե': "e",
	'Զ': "Z", 'զ': "z",
	'Է': "E", 'է': "e",
	'Ը': "Y", 'ը': "y",
	'Թ': "T'", 'թ': "t'",
	'Ժ': "Zh", 'ժ': "zh",
	'Ի': "I", 'ի': "i",
	'Լ': "L", 'լ': "l",
	'Խ': "Kh", 'խ': "kh",
	'Ծ': "Ts", 'ծ': "ts",
	'Կ': "K", 'կ': "k",
	'Հ': "H", 'հ': "h",
	'Ձ': "Dz", 'ձ': "dz",
	'Ղ': "Gh", 'ղ': "gh",
	'Ճ': "Tch", 'ճ': "tch",
	'Մ': "M", 'մ': "m",
	'Յ': "Y", 'յ': "y",
	'Ն': "N", 'ն': "n",
	'Շ': "Sh", 'շ': "sh",
	'Ո': "Vo", 'ո': "vo",
	'Չ': "Ch'", 'չ': "ch'",
	'Պ': "P", 'պ': "p",
	'Ջ': "J", 'ջ': "j",
	'Ռ': "R", 'ռ': "r",
	'Ս': "S", 'ս': "s",
	'Վ': "V", 'վ': "v",
	'Տ': "T", 'տ': "t",
	'Ր': "R", 'ր': "r",
	'Ց': "Ts'", 'ց': "ts'",
	'Ւ': "V", 'ւ': "v",
	'Փ': "P'", 'փ': "p'",
	'Ք': "Q", 'ք': "q",
	'Օ': "O", 'օ': "o",
	'Ֆ': "F", 'ֆ': "f",
	// ՙ, ՚, ՛, ՜, ՝, ՞, and ՟ are punctuation marks and do not have Latin equivalents
}

// Basic mapping from Lao characters to the Latin alphabet.
var laoToLatinMap = map[string]string{
	// Consonants
	"ກ": "k", "ຂ": "kh", "ຄ": "kh", "ງ": "ng",
	"ຈ": "ch", "ສ": "s", "ຊ": "x", "ຍ": "ny",
	"ດ": "d", "ຕ": "t", "ຖ": "th", "ທ": "th", "ນ": "n",
	"ບ": "b", "ປ": "p", "ຜ": "ph", "ຝ": "f", "ພ": "ph", "ຟ": "f", "ມ": "m",
	"ຢ": "y", "ຣ": "r", "ລ": "l", "ວ": "w", "ຫ": "h",
	"ອ": "o", "ຮ": "h",

	// Vowels
	"ະ": "a", "ັ": "a", "າ": "ā", "ຳ": "am", "ິ": "i", "ີ": "ī",
	"ຶ": "u", "ື": "ū", "ຸ": "u", "ູ": "ū", "ົ": "o", "ຼ": "l",
	"ເ": "e", "ແ": "ē", "ໂ": "o", "ໃ": "ai", "ໄ": "ai",

	// Tone marks
	"່": "", "້": "", "໊": "", "໋": "",

	// Other symbols
	"໌": "", "ໍ": "",

	// Numbers
	"໐": "0", "໑": "1", "໒": "2", "໓": "3", "໔": "4",
	"໕": "5", "໖": "6", "໗": "7", "໘": "8", "໙": "9",
}

var ahmaricReplacements = map[string]string{
	"ሀ": "hä", "ለ": "lä", "ሐ": "hä", "መ": "mä", "ሠ": "sä", "ረ": "rä", "ሰ": "sä", "ሸ": "šä",
	"ቀ": "qä", "በ": "bä", "ተ": "tä", "ቸ": "čä", "ኀ": "hä", "ነ": "nä", "ኘ": "ñä", "አ": "ʾä",
	"ከ": "kä", "ኸ": "hä", "ወ": "wä", "ዐ": "ʾä", "ዘ": "zä", "ዠ": "žä", "የ": "yä", "ደ": "dä",
	"ጀ": "ǧä", "ገ": "gä", "ጠ": "t'ä", "ጨ": "č'ä", "ጰ": "p'ä", "ጸ": "s'ä", "ፀ": "s'ä",
	"ፈ": "fä", "ፐ": "pä", "ሁ": "hu", "ሉ": "lu", "ሑ": "hu", "ሙ": "mu", "ሡ": "su", "ሩ": "ru",
	"ሱ": "su", "ሹ": "šu", "ቁ": "qu", "ቡ": "bu", "ቱ": "tu", "ቹ": "ču", "ኁ": "hu", "ኑ": "nu",
	"ኙ": "ñu", "ኡ": "ʾu", "ኩ": "ku", "ኹ": "hu", "ዉ": "wu", "ዑ": "ʾu", "ዙ": "zu", "ዡ": "žu",
	"ዩ": "yu", "ዱ": "du", "ጁ": "ǧu", "ጉ": "gu", "ጡ": "t'u", "ጩ": "č'u", "ጱ": "p'u", "ጹ": "s'u",
	"ፁ": "s'u", "ፉ": "fu", "ፑ": "pu", "ሂ": "hi", "ሊ": "li", "ሒ": "hi", "ሚ": "mi", "ሢ": "si",
	"ሪ": "ri", "ሲ": "si", "ሺ": "ši", "ቂ": "qi", "ቢ": "bi", "ቲ": "ti", "ቺ": "či", "ኂ": "hi",
	"ኒ": "ni", "ኚ": "ñi", "ኢ": "ʾi", "ኪ": "ki", "ኺ": "hi", "ዊ": "wi", "ዒ": "ʾi", "ዚ": "zi",
	"ዢ": "ži", "ዪ": "yi", "ዲ": "di", "ጂ": "ǧi", "ጊ": "gi", "ጢ": "t'i", "ጪ": "č'i", "ጲ": "p'i",
	"ጺ": "s'i", "ፂ": "s'i", "ፊ": "fi", "ፒ": "pi", "ሃ": "ha", "ላ": "la", "ሓ": "ha", "ማ": "ma",
	"ሣ": "sa", "ራ": "ra", "ሳ": "sa", "ሻ": "ša", "ቃ": "qa", "ባ": "ba", "ታ": "ta", "ቻ": "ča",
	"ኃ": "ha", "ና": "na", "ኛ": "ña", "ኣ": "ʾa", "ካ": "ka", "ኻ": "ha", "ዋ": "wa", "ዓ": "ʾa",
	"ዛ": "za", "ዣ": "ža", "ያ": "ya", "ዳ": "da", "ጃ": "ǧa", "ጋ": "ga", "ጣ": "t'a", "ጫ": "č'a",
	"ጳ": "p'a", "ጻ": "s'a", "ፃ": "s'a", "ፋ": "fa", "ፓ": "pa", "ሄ": "he", "ሌ": "le", "ሔ": "he",
	"ሜ": "me", "ሤ": "se", "ሬ": "re", "ሴ": "se", "ሼ": "še", "ቄ": "qe", "ቤ": "be", "ቴ": "te",
	"ቼ": "če", "ኄ": "he", "ኔ": "ne", "ኜ": "ñe", "ኤ": "ʾe", "ኬ": "ke", "ኼ": "he", "ዌ": "we",
	"ዔ": "ʾe", "ዜ": "ze", "ዤ": "že", "ዬ": "ye", "ዴ": "de", "ጄ": "ǧe", "ጌ": "ge", "ጤ": "t'e",
	"ጬ": "č'e", "ጴ": "p'e", "ጼ": "s'e", "ፄ": "s'e", "ፌ": "fe", "ፔ": "pe", "ህ": "hə", "ል": "lə",
	"ሕ": "hə", "ም": "mə", "ሥ": "sə", "ር": "rə", "ስ": "sə", "ሽ": "šə", "ቅ": "qə", "ብ": "bə",
	"ት": "tə", "ች": "čə", "ኅ": "hə", "ን": "nə", "ኝ": "ñə", "እ": "ʾə", "ክ": "kə", "ኽ": "hə",
	"ው": "wə", "ዕ": "ʾə", "ዝ": "zə", "ዥ": "žə", "ይ": "yə", "ድ": "də", "ጅ": "ǧə", "ግ": "gə",
	"ጥ": "t'ə", "ጭ": "č'ə", "ጵ": "p'ə", "ጽ": "s'ə", "ፅ": "s'ə", "ፍ": "fə", "ፕ": "pə", "ሆ": "ho",
	"ሎ": "lo", "ሖ": "ho", "ሞ": "mo", "ሦ": "so", "ሮ": "ro", "ሶ": "so", "ሾ": "šo", "ቆ": "qo",
	"ቦ": "bo", "ቶ": "to", "ቾ": "čo", "ኆ": "ho", "ኖ": "no", "ኞ": "ño", "ኦ": "ʾo", "ኮ": "ko",
	"ኾ": "ho", "ዎ": "wo", "ዖ": "ʾo", "ዞ": "zo", "ዦ": "žo", "ዮ": "yo", "ዶ": "do", "ጆ": "ǧo",
	"ጎ": "go", "ጦ": "t'o", "ጮ": "č'o", "ጶ": "p'o", "ጾ": "s'o", "ፆ": "s'o", "ፎ": "fo", "ፖ": "po",
	"ጐ": "gu", "ጓ": "gä", "ሏ": "wa", "ሟ": "ma", "ቷ": "ta",
	"፤": ":", "፡": " ",
	"።": ".", "፩": "1", "፪": "2", "፫": "3", "፬": "4", "፭": "5", "፮": "6", "፯": "7", "፰": "8",
	"፱": "9", "፲": "10",
}

// Basic mapping from Khmer characters to Latin alphabet. This is very simplified and needs to be expanded.
var khmerToLatinMap = map[string]string{
	// Consonants
	"ក": "k", "ខ": "kh", "គ": "k", "ឃ": "kh", "ង": "ng",
	"ច": "ch", "ឆ": "chh", "ជ": "ch", "ឈ": "chh", "ញ": "nh",
	"ដ": "d", "ឋ": "th", "ឌ": "d", "ឍ": "th", "ណ": "n",
	"ត": "t", "ថ": "th", "ទ": "t", "ធ": "th", "ន": "n",
	"ប": "b", "ផ": "ph", "ព": "p", "ភ": "ph", "ម": "m",
	"យ": "y", "រ": "r", "ល": "l", "វ": "v", "ឝ": "s",
	"ឞ": "sa", "ស": "s", "ហ": "h", "ឡ": "l", "អ": "'",

	// Independent vowels
	"ឣ": "a", "ឤ": "aa", "ឥ": "i", "ឦ": "ii", "ឧ": "u",
	"ឨ": "uk", "ឩ": "uu", "ឪ": "uuv", "ឫ": "ry", "ឬ": "ryy",
	"ឭ": "ly", "ឮ": "lyy", "ឯ": "e", "ឰ": "ai", "ឱ": "oo",
	"ឲ": "au", "ឳ": "au",

	// Dependent vowels
	"ា": "a", "ិ": "i", "ី": "ii", "ឹ": "u", "ឺ": "uu",
	"ុ": "u", "ូ": "uu", "ួ": "uo", "ើ": "ae", "ឿ": "ie",
	"ៀ": "e", "េ": "ae", "ែ": "ai", "ៃ": "ai", "ោ": "ao",
	"ៅ": "au", "ំ": "am", "ះ": "ah",

	// Subscripts
	"្": "", // This is a subscript marker; actual subscript consonants would need additional rules

	// Diacritics and other marks
	"ៈ": "", // Symbol to indicate duplicate consonant
	"៉": "", // Indicates change in consonant sound
	"៊": "", // Another mark for consonant sound change
	"់": "", // Tone mark
	"៌": "", // Tone mark
	"៍": "", // Cancel previous diacritic
	"៎": "", // Rare mark, usage varies
	"៏": "", // Rare mark, usage varies
	"័": "", // Series marker
	"៑": "", // Obscure diacritic

	// Numbers
	"០": "0", "១": "1", "២": "2", "៣": "3", "៤": "4",
	"៥": "5", "៦": "6", "៧": "7", "៨": "8", "៩": "9",
}

// Basic mapping from Thai characters to Latin alphabet. This is very simplified and needs to be expanded.
var thaiToLatinMap = map[string]string{
	// Consonants
	"ก": "k", "ข": "kh", "ฃ": "kh", "ค": "kh", "ฅ": "kh", "ฆ": "kh",
	"ง": "ng", "จ": "j", "ฉ": "ch", "ช": "ch", "ซ": "s", "ฌ": "ch",
	"ญ": "y", "ฎ": "d", "ฏ": "t", "ฐ": "th", "ฑ": "th", "ฒ": "th",
	"ณ": "n", "ด": "d", "ต": "t", "ถ": "th", "ท": "th", "ธ": "th",
	"น": "n", "บ": "b", "ป": "p", "ผ": "ph", "ฝ": "f", "พ": "ph",
	"ฟ": "f", "ภ": "ph", "ม": "m", "ย": "y", "ร": "r", "ล": "l",
	"ว": "w", "ศ": "s", "ษ": "s", "ส": "s", "ห": "h", "ฬ": "l",
	"อ": "o", "ฮ": "h",

	// Vowels and vowel combinations
	"ะ": "a", "ั": "a", "า": "a", "ำ": "am", "ิ": "i", "ี": "i",
	"ึ": "ue", "ื": "ue", "ุ": "u", "ู": "u", "เ": "e", "แ": "ae",
	"โ": "o", "ใ": "ai", "ไ": "ai",

	// Tone marks
	"่": "", "้": "", "๊": "", "๋": "",

	// Other symbols
	"ฯ": ".", "ๆ": "(repetition)", "์": "", "ํ": "", "๏": "section",
	"๐": "0", "๑": "1", "๒": "2", "๓": "3", "๔": "4",
	"๕": "5", "๖": "6", "๗": "7", "๘": "8", "๙": "9",
}

// The following functions are responsible for latinizations of various scripts; it is specified in the name.

func LatinizeCyrillic(text string) string {
	var outputString string

	for _, char := range text {
		value, found := cyrillicToLatin[string(char)]
		if found {
			outputString += value
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeGreek(text string) string {
	var outputString string

	for _, char := range text {
		value, found := greekToLatin[string(char)]
		if found {
			outputString += value
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeChinese(text string, data map[string][]string) string {
	var outputString string
	for _, char := range text {
		value, found := data[string(char)]
		if found {
			outputString += value[0]
		} else {
			outputString += string(char)
		}
	}
	return outputString
}

func LatinizeJapanese(text string) string {
	var output string
	url := "https://japonesbasico.com/furigana/procesa.php"
	// This is the data that will be sent in the request body:
	data := []byte(fmt.Sprintf(`{"conversion":"romaji", "japaneseText":"%s", "lang":"en"}`, text))

	// Make the HTTP POST request
	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err.Error()
	}
	defer response.Body.Close()

	// Check if the response status code is 200 OK
	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("Unexpected status code: %d\n", response.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response body: %s", err.Error())
	}
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err.Error()
	}

	// Find and print the value inside the second <td> element
	output = findSecondTdValue(doc)
	return output
}

// Latinization of Arabic

func LatinizeArabic(input string) string {

	// Replace Arabic characters with Latin characters
	for arChar, latinChar := range arabicToLatin {
		input = strings.ReplaceAll(input, arChar, latinChar)
	}

	return input
}

// Latinization of hindi

func LatinizeHindi(input string) string {
	// Replace Hindi characters with Latin characters
	for hindiChar, latinChar := range HindiReplacements {
		input = strings.ReplaceAll(input, hindiChar, latinChar)
	}

	return input
}

// Latinization of farsi (persian)

func LatinizePersian(input string) string {
	// Replace Farsi characters with Latin characters
	for farsiChar, latinChar := range FarsiReplacements {
		input = strings.ReplaceAll(input, farsiChar, latinChar)
	}

	return input
}

func LatinizeBurmese(burmeseText string) string {
	var latinText strings.Builder

	for _, char := range burmeseText {
		latinChar, ok := burmeseToLatinMap[string(char)]
		if ok {
			latinText.WriteString(latinChar)
		} else {
			latinText.WriteRune(char) // Keep the character as is if no mapping is found
		}
	}

	return latinText.String()
}

// LatinizeLao converts Lao text to a simplified Latin representation.
func LatinizeLao(laoText string) string {
	var latinText strings.Builder

	for _, char := range laoText {
		latinChar, ok := laoToLatinMap[string(char)]
		if ok {
			latinText.WriteString(latinChar)
		} else {
			latinText.WriteRune(char) // Keep the character as is if no mapping is found
		}
	}

	return latinText.String()
}

// LatinizeKhmer converts Khmer text to a simplified Latin representation.
func LatinizeKhmer(khmerText string) string {
	var latinText strings.Builder

	for _, char := range khmerText {
		latinChar, ok := khmerToLatinMap[string(char)]
		if ok {
			latinText.WriteString(latinChar)
		} else {
			latinText.WriteRune(char) // Keep the character as is if no mapping is found
		}
	}

	return latinText.String()
}

func LatinizeThai(thaiText string) string {
	var latinText strings.Builder

	for _, char := range thaiText {
		latinChar, ok := thaiToLatinMap[string(char)]
		if ok {
			latinText.WriteString(latinChar)
		} else {
			latinText.WriteRune(char) // Keep the character as is if no mapping is found
		}
	}

	return latinText.String()
}

// Function to latinize hebrew text.

func LatinizeHebrew(input string) string {
	sym := input
	sym = strings.ReplaceAll(sym, "א", "a")
	sym = strings.ReplaceAll(sym, "ב", "b")
	sym = strings.ReplaceAll(sym, "ג", "g")
	sym = strings.ReplaceAll(sym, "ד", "d")
	sym = strings.ReplaceAll(sym, "ה", "h")
	sym = strings.ReplaceAll(sym, "ו", "v")
	sym = strings.ReplaceAll(sym, "ז", "z")
	sym = strings.ReplaceAll(sym, "ח", "h")
	sym = strings.ReplaceAll(sym, "ט", "t")
	sym = strings.ReplaceAll(sym, "י", "y")
	sym = strings.ReplaceAll(sym, "ך", "k")
	sym = strings.ReplaceAll(sym, "כ", "k")
	sym = strings.ReplaceAll(sym, "ל", "l")
	sym = strings.ReplaceAll(sym, "ם", "m")
	sym = strings.ReplaceAll(sym, "מ", "m")
	sym = strings.ReplaceAll(sym, "ן", "n")
	sym = strings.ReplaceAll(sym, "נ", "n")
	sym = strings.ReplaceAll(sym, "ס", "s")
	sym = strings.ReplaceAll(sym, "ע", "'e")
	sym = strings.ReplaceAll(sym, "ף", "p")
	sym = strings.ReplaceAll(sym, "פ", "p")
	sym = strings.ReplaceAll(sym, "ץ", "ts")
	sym = strings.ReplaceAll(sym, "צ", "ts")
	sym = strings.ReplaceAll(sym, "ק", "q")
	sym = strings.ReplaceAll(sym, "ר", "r")
	sym = strings.ReplaceAll(sym, "ש", "sh")
	sym = strings.ReplaceAll(sym, "ת", "t")
	sym = strings.ReplaceAll(sym, "ב", "b")
	sym = strings.ReplaceAll(sym, "כ", "k")
	sym = strings.ReplaceAll(sym, "פ", "p")
	sym = strings.ReplaceAll(sym, "ת", "t")
	sym = strings.ReplaceAll(sym, "ו", "u")
	sym = strings.ReplaceAll(sym, "ו", "v")
	sym = strings.ReplaceAll(sym, "וֹ", "o")
	sym = strings.ReplaceAll(sym, "ָ", "a")
	sym = strings.ReplaceAll(sym, "ַ", "a")
	sym = strings.ReplaceAll(sym, "ּ", "i")
	sym = strings.ReplaceAll(sym, "ײ", "i")
	sym = strings.ReplaceAll(sym, "װ", "y")
	sym = strings.ReplaceAll(sym, "ױ", "yi")
	sym = strings.ReplaceAll(sym, "ֿ", "a")
	sym = strings.ReplaceAll(sym, "־", "")
	sym = strings.ReplaceAll(sym, "\n", "\n")
	return sym
}

func LatinizeGeorgian(input string) string {
	var result strings.Builder
	for _, char := range input {
		if latinChar, ok := georgianToLatin[char]; ok {
			result.WriteString(latinChar)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func LatinizeArmenian(input string) string {
	var result strings.Builder
	for _, char := range input {
		if latinChar, ok := armenianToLatin[char]; ok {
			result.WriteString(latinChar)
		} else {
			// If the character does not have a mapping, just add it as is.
			result.WriteRune(char)
		}
	}
	return result.String()
}

func LatinizeAhmaric(input string) string {
	for old, new := range ahmaricReplacements {
		input = strings.Replace(input, old, new, -1)
	}
	return input
}

func findSecondTdValue(n *html.Node) string {
	var result string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "td" {
			// Check if it is the second <td> element
			if node.NextSibling == nil {
				// Get the text content of the second <td>
				result = getTextContent(node.FirstChild)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return result
}

func getTextContent(n *html.Node) string {
	var result string

	if n != nil {
		if n.Type == html.TextNode {
			result += n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result += getTextContent(c)
		}
	}

	return result
}

/* InitHanzi function:
This function is responsible for the "translation" of the data inside the hanzi.json file (i.e the map
that associates chinese characters to their pinyin romanizations) inside a Go map object.
It returns the map we obtain.
*/

func InitHanzi(hanzi []byte) map[string][]string {
	// Initialize the map we're going to return
	var data map[string][]string

	// Use json.Unmarshal to decode the json into the map
	err := json.Unmarshal(hanzi, &data)
	// If there's an error print it out to the console.
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
	// return the data map.
	return data
}

/* LatinizeText function
This function is responsible for the latinization of a certain portion of a text.
Input: 1) the text we want to latinize (string)
2) data, which is the map that contains the latinizations in pinyin of the chinese hanzi characters.
3) language; which is the language we want to latinize (i.e the language we're currently studying).
*/

func LatinizeText(text string, data map[string][]string, language string) string {
	// Check what language we're studying
	switch language {
	case "chinese":
		// If it's chinese latinize the chinese into pinyin
		return LatinizeChinese(text, data)
	case "russian", "serbian", "mongolian", "belarusian", "ukrainian", "bulgarian", "kazakh":
		// If it's one of these languages, then the script used is cyrillic: latinize accordingly.
		return LatinizeCyrillic(text)
	case "greek":
		// Greek script latinization
		return LatinizeGreek(text)
	case "japanese":
		return LatinizeJapanese(text)
	case "korean":
		return LatinizeKorean(text)
	case "arabic":
		return LatinizeArabic(text)
	case "hindi":
		return LatinizeHindi(text)
	case "persian":
		return LatinizePersian(text)
	case "hebrew":
		return LatinizeHebrew(text)
	case "burmese":
		return LatinizeBurmese(text)
	case "lao":
		return LatinizeLao(text)
	case "khmer":
		return LatinizeKhmer(text)
	case "thai":
		return LatinizeThai(text)
	case "armenian":
		return LatinizeArmenian(text)
	case "georgian":
		return LatinizeGeorgian(text)
	case "tigre", "tigrinya", "ahmaric":
		return LatinizeAhmaric(text)
	default:
		// else just returns the text; I still need to cover a lot of non-latin scripts.
		return text

	}
}
