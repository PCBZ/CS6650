# Crashing and Recovering Mappers in MapReduce
In this part of the midterm, I enhanced my MapReduce implementation to handle mapper failures gracefully. I created a "malfunctional" mapper that return 500 error during execution to simulate real-world failures. 

For an example file, I count it on a local script, showing results:
```bash
Downloading text from: s3://mapreduce-experiment-975050147762/50MB-TXT-FILE.txt
Downloaded 52433149 characters
Word counting completed in 1.84 seconds

Word Count Results:
Total words: 6883544
Unique words: 311
```