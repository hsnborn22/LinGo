from math import log
from collections import defaultdict
from enum import Enum
from word_breaker.Rabbit import zg2uni
from word_breaker.myparser import MyParser

import os

class WordSegment:
    # Word Segmentation Ways
    SegmentationMethod = Enum('SegmentationMethod', 'all_possible_combination sub_word_possibility')

    # __file__ refers to the file settings.py
    APP_ROOT = os.path.dirname(os.path.abspath(__file__))  # refers to application_top
    APP_DICTIONARY = os.path.join(APP_ROOT, 'dictionary')
    APP_CORPUS = os.path.join(APP_ROOT, 'corpus')

    # this will instead read the file into a python list.
    _dict_words = open(os.path.join(APP_DICTIONARY, 'dict-words.txt'), 'r', encoding='utf-8').read().splitlines()
    _stop_words = open(os.path.join(APP_DICTIONARY, 'stopwords.txt'), 'r', encoding='utf-8').read().splitlines()

    _mypos_corpus = open(os.path.join(APP_CORPUS, 'mypos-dver.1.0.cword.txt'), 'r', encoding='utf-8').read()
    _total_unigram_count = 341685  # total words count (including the factor of compound words)
    _total_bigram_count = 170843  # total words count (including the factor of compound words)

    _not_found_words = set()
    _found_words = set()

    _possible_combos = []
    _maxlen = 6

    m = MyParser()

    # check word include in dictionary
    def _check_in_dicts(self, word):
        found = True;

        if word in self._found_words:
            # if word is already in found_words list
            found = True
        elif word in self._not_found_words:
            # if word is already in not found_words list
            found = False;
        elif not (word in self._dict_words or
                  word in self._stop_words):
            # if word is not in the dictionary
            # we add to not_found_words list
            self._not_found_words.add(word)
            found = False
        else:
            # nothing else, we assume this is the valid word
            self._found_words.add(word)

        return found;

    # simple left to right maximum longest matching
    def _left_to_right_segment(self, seq, maxlen):
        length = len(seq)
        offset = 0
        combo = []

        # first Left to Right segmentation
        while length > 0:
            for i in range(maxlen, 0, -1):
                # make the chunk
                chunk = offset + i

                # create the word from the chunk
                word = ''.join(seq[offset:chunk])

                # check in dictionary
                if self._check_in_dicts(word) or i == 1:
                    # found word
                    combo.append(word)
                    offset += i
                    length -= i
                    break

        return combo

    # make all possible word combinations
    def _make_combinations(self, seq, maxlen):
        # memo is a dict of {length_of_last_word: list_of_combinations}
        memo = defaultdict(list)

        # put the first character into the memo
        memo[1] = [[seq[0]]]

        seq_iter = iter(seq)
        next(seq_iter)  # skip the first character
        last_index = len(seq) - 2
        for index, char in enumerate(seq_iter):
            new_memo = defaultdict(list)

            # iterate over the memo and expand it
            for wordlen, combos in memo.items():
                if not combos:
                    continue

                # add the current character as a separate word
                new_memo[1].extend(combo + [char] for combo in combos)

                # if the maximum word length isn't reached yet, add a character to the last word
                if wordlen < maxlen:
                    longest_word = combos[0][-1]
                    word = combos[0][-1] + char

                    new_memo[wordlen + 1] = newcombos = []
                    for combo in combos:
                        combo[-1] = word  # overwrite the last word with a longer one
                        newcombos.append(combo)

                if index == last_index or wordlen + 1 == maxlen:
                    if word and not self._check_in_dicts(word):
                        word = ''
                        del new_memo[wordlen + 1]

                if longest_word and not longest_word in seq:
                    if not self._check_in_dicts(longest_word):
                        longest_word = ''
                        del new_memo[1][len(combos) * -1:]

            memo = new_memo

        # flatten the memo into a list and return it
        combos = []
        for combo in memo.values():
            combos.extend(combo)

        return combos

    # calulate mutual information of two syllables
    def _cal_mutual_info(self, sya1, sya2):
        # calculating unigram probability
        occurence_of_syllable_1 = self._mypos_corpus.count(sya1)
        probability_of_syllable_1 = occurence_of_syllable_1 / self._total_unigram_count

        occurence_of_syllable_2 = self._mypos_corpus.count(sya2)
        probability_of_syllable_2 = occurence_of_syllable_2 / self._total_unigram_count

        # calculating bigram probability
        occurence_of_bigram = self._mypos_corpus.count(sya1 + sya2)
        probability_of_bigram = occurence_of_bigram / self._total_bigram_count

        if probability_of_bigram == 0:
            return 0

        return log(probability_of_bigram / float(probability_of_syllable_1 * probability_of_syllable_2), 2)

    def filter_minimum_combination(self, solutions):
        # retrieving minimum number of merged words
        min_filtering_count = min(map(len, solutions))
        min_filtering_solutions = []
        for solution in solutions:
            if len(solution) == min_filtering_count:
                min_filtering_solutions.append(solution)

        return min_filtering_solutions

    def _calculate_sentence_collocation_strength(self, filtering_solutions):
        syllable_sents = []

        for solution in filtering_solutions:
            syllable_sents.append([self.m.syllable(word) for word in solution])

        syllable_collocation_strength = []
        for sent_index, syllable_sent in enumerate(syllable_sents):
            sentence_collocation_strength = 0
            for index, syllable_word in enumerate(syllable_sent):
                i = 0
                word_mutual_info = 0
                if (len(syllable_word) > 1):
                    # calculate only if syllable_word is at least longer than unigram
                    # calculating positive strength
                    while (i < len(syllable_word)):
                        # calculate positive collocation strength
                        if (i < len(syllable_word) - 1):
                            word_mutual_info += self._cal_mutual_info(syllable_word[i], syllable_word[i + 1])
                        i += 1

                    # calculating left negative strength
                    # check previous item exists
                    if index - 1 >= 0:
                        # calculate negative left collocation strength
                        left_last_syllable = syllable_sent[index - 1][-1]
                        word_mutual_info -= self._cal_mutual_info(left_last_syllable, syllable_word[0])

                    # calculating right negative strength
                    # check next item exists
                    if (index + 1 < len(syllable_sent)):
                        # calculate negative right collocation strength
                        right_last_syllable = syllable_sent[index + 1][0]
                        word_mutual_info -= self._cal_mutual_info(syllable_word[-1], right_last_syllable)

                sentence_collocation_strength += word_mutual_info

            syllable_collocation_strength.append([sentence_collocation_strength, filtering_solutions[sent_index]])

        return syllable_collocation_strength

    def _make_sub_word_combinations(self, input, combo, shortest_length, start_word_position=0, pointer=0):
        # shortest length = shortest word combinations count
        for n in range(start_word_position, shortest_length):
            result = combo[0:n]
            word = combo[n]
            seq = self.m.syllable(word)

            offset = 0
            maxlen = len(seq) - 1

            for i in range(maxlen, 0, -1):
                chunk = offset + i
                word = ''.join(seq[offset:chunk])

                # check in dictionary
                if self._check_in_dicts(word):
                    # found word
                    result.append(word)
                    result.extend(self._left_to_right_segment(input[pointer + i:], self._maxlen))

                    if (len(result) <= shortest_length):
                        self._possible_combos.append(result)
                        cur_pointer = pointer + i
                        self._make_sub_word_combinations(input, result, shortest_length, n + 1, cur_pointer)

                    break

            pointer += len(seq)

    def break_words(self, input, segmentation_method):
        # Breaking up words into syllables
        input = self.m.syllable(input)

        # Compose all the segmentations
        if (segmentation_method == self.SegmentationMethod.all_possible_combination):
            self._possible_combos = self._make_combinations(input, self._maxlen)
        elif (segmentation_method == self.SegmentationMethod.sub_word_possibility):
            combo = self._left_to_right_segment(input, self._maxlen)
            self._possible_combos.append(combo)
            self._make_sub_word_combinations(input, combo, len(combo))

        min_filtered_combos = self.filter_minimum_combination(self._possible_combos)
        if len(min_filtered_combos) > 1:
            syllable_collocation_strengths = self._calculate_sentence_collocation_strength(min_filtered_combos)
            strongest = 0
            strongest_sentence = ''
            # comb => [strength , sentence]
            for comb in syllable_collocation_strengths:
                strength = comb[0]
                if (strength > strongest):
                    strongest = strength
                    strongest_sentence = comb[1]

            return strongest_sentence
        else:
            return min_filtered_combos[0]

    def normalize_break(self, input_text, encoding, segmentation_method=SegmentationMethod.all_possible_combination):
        # if it is zawgyi, converts to Unicode
        if (encoding == "zawgyi"):
            input_text = zg2uni(input_text)

        # normalize the input text
        input_text = input_text.replace(u" ", "")
        inputs = input_text.split("။")

        outputs = []

        for input in inputs:
            if input:
                segmented_sentence = self.break_words(input, segmentation_method)
                outputs.append(segmented_sentence)

        return outputs
