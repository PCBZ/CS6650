#!/usr/bin/env python3

import re
import urllib.request
from collections import Counter
import time
from typing import Tuple

class WordCount:
    def download_text(self, url):
        """Download text from URL"""
        try:
            with urllib.request.urlopen(url) as response:
                return response.read().decode('utf-8')
        except Exception as e:
            print(f"Error downloading file: {e}")
            return None

    def count_words(self, text):
        """Count word frequencies in text"""
        # Use regex to extract words (letters, numbers, apostrophes)
        words = re.findall(r"[a-zA-Z0-9']+", text)
        
        # Convert to lowercase and remove leading/trailing apostrophes
        cleaned_words = []
        for word in words:
            word = word.lower().strip("'")
            if word:
                cleaned_words.append(word)
        
        return Counter(cleaned_words)

def main():
    # url = "https://raw.githubusercontent.com/teropa/nlp/master/resources/corpora/gutenberg/shakespeare-hamlet.txt"
    url = "https://www.gutenberg.org/files/100/100-0.txt"

    print("Downloading Hamlet text...")
    counter = WordCount()

    text = counter.download_text(url)
    
    if text is None:
        return
    
    print(f"Downloaded {len(text)} characters")

    start_time = time.time()

    # Count words
    counter = counter.count_words(text)

    duration = time.time() - start_time
    print(f"Word counting completed in {duration:.2f} seconds")
    
    # Print results
    print(f"\nWord Count Results:")
    print(f"Total words: {sum(counter.values())}")
    print(f"Unique words: {len(counter)}")

if __name__ == "__main__":

    main()