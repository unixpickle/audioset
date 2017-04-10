# Downloading AudioSet

This is a tool for fetching data from [AudioSet](https://research.google.com/audioset/download.html), a dataset of audio events.

# Dependencies

 * bash
 * ffmpeg
 * youtube-dl
 * gzip

# Usage

Simply pipe a CSV file from the dataset into download.sh. For example:

```
$ cat eval_segments.csv | ./download.sh
```

The WAV files will automatically be downloaded into the current directory and compressed with gzip.
