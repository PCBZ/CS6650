#!/usr/bin/env python3

import re
import urllib.request
from collections import Counter
import time
from typing import Tuple
import boto3

class WordCount:
    def download_text(self, url):
        """Download text from URL"""
        try:
            if url.startswith('s3://'):
                # Parse S3 URL
                parts = url[5:].split('/', 1)
                bucket = parts[0]
                key = parts[1]
                
                # Use boto3 to download from S3
                s3_client = boto3.client('s3', region_name='us-west-2')
                response = s3_client.get_object(Bucket=bucket, Key=key)
                return response['Body'].read().decode('utf-8')
            else:
                # HTTP/HTTPS URL
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
    # S3 URL example
    url = "s3://mapreduce-experiment-975050147762/50MB-TXT-FILE.txt"
    # HTTP URL example
    # url = "https://www.gutenberg.org/files/100/100-0.txt"

    print("Downloading text from:", url)
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