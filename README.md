# LinGo
LinGo is a terminal based language aquisiton application, which allows you to learn languages while reading texts. The application allows you to load texts, read them, mark words with different colors based on your knowledge of them, listen the pronunciation of words on the spot and get a translation of each word without leaving the application; all the words you marked will then be saved locally on your machine and can be easily imported into anki or memrise flashcards. The application is entirely keyboard-based, making it way faster to use compared to the other alternatives. 

Below some tests to show how the application works.

https://github.com/hsnborn22/LinGo/assets/146590127/bd237dd4-11a7-4c0a-b9fe-fadccd8d4547


https://github.com/hsnborn22/LinGo/assets/146590127/c55b0472-71cd-4615-bed0-7f5786bf9309

## Motivation:
The motivation for the application comes from multiple sources; on a conceptual level, the application is based on the technique of language aquisition via text, a technique which is deemed to be highly effective by many linguists and polyglots. However, when you read a text in a language you're still learning, new words that you don't know and you don't know the pronunciation of are quite common: these occurences force you to move your attention away from the text, go to an online (or physical) dictionary, look up the meaning of the word, listen to its pronunciation exc. With this application (which really started out as a tool for myself precisely to speed up this process) you can do all of this without moving away from the text you're reading, making the process of learning by aquisition faster, easier and more enjoyable.

Another necessity which we could encounter frequently is that of remembering these new words we encountered and then maybe inserting them in some flashcard system in order to master them.  
Let's envision this process if we had to absorb a language by reading physical books via an "algorithmic" procedure:

Reading --> encounter new word --> look up meaning and pronunciation elsewhere --> sign on the physical book that we do not know this word (for example by underlining it --> when we are done with our reading session, find all those words and manually make them into an anki or memrise flashcard set.

Wow, it takes really a lot of time! Well, I have good news for you: this application automates all of this for you. While you're reading, you can mark a word as known by pressing the key "3" (and the app will mark it as green), as semi-known by pressing the key "2" (and the app will mark the word as yellow) and as not known by pressing the key "1" (and the app will mark the word as red). By pressing the key 0 you can decide to ignore a word (by default, if you never encountered a word before, the value is set to 0 so you can decide if you know it or not).

All this information about words and how well you know them are then saved locally on your computer on a json file, so that when you open another text in the same language, and that text has a word that you already encountered in another text, it will still have the same value that you assigned to it. 

To make flashcards out of the words that you do not know, you can then just press the "f" key in the language menu or directly while you're reading a text, and a new file "dictionary.txt" will be created in the folder of the language you're studying. This file is formatted in such a way that can be imported into flashcards in both anki and memrise, since the words will be stored like this:

word1,translation1

word2,translation2

...

Summing everything up, the goal of the app is to provide a lightweight, fast to use, user-friendly terminal-based solution to language aquisition by texts, that gives off "VIM" vibes. I'm sure if you're a software developer who likes both VIM and learning languages you will really like this. 

## Languages supported:

At the current stage the application supports virtually all languages that are "tokenized with spaces", i.e where words are separated by spaces (almost all european languages are like this, as well as many asian languages such as indonesian, tagalog,mongolian, kazakh exc.). The app also supports mandarin chinese (both simplified and traditional) both for tts and for in-text translations. Languages such as arabic, korean and japanese are currently not supported yet.
If you want to study a less popular language which is still supported (maybe a conlang) such as lojban, toki pona or even your own conlang, a procedure to add it to the app will be explained after (don't worry it's very easy). However note that tts or translation might not be available. 