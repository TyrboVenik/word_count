# Word Count

CLI that calculates the top words used in the provided articles.

## Table of Contents

- [Usage](#usage)
- [Description](#description)
- [Contributing](#contributing)
- [Comments](#comments)

## Usage
Requirements: Golang

    go mod download
    make build
    
    make run
    or 
    ./bin/wc -rps 2 -workers 8 -batch-size 4000
    
    more info 
    ./bin/wc --help

     Usage of ./bin/wc:
     -articles-url string
             articles file url (default "https://drive.google.com/uc?id=1TF4RPuj8iFwpa-lyhxG67V8NDlktmTGi&export=download")
     -bank-of-words-url string
             bank of words file url (default "https://raw.githubusercontent.com/dwyl/english-words/master/words.txt")
     -batch-size int
             process first <batch-size> articles. 0 - process all
     -n int
             amount of top words to be shown (default 10)
     -progress int
             show progress every <progress> requests. 0 - hide progress (default 100)
     -rps int
             rps limit (default 2)
     -workers int
             parallel requests limit (default 8)

## Contributing

Requirements: Golang and [golangci-lint](https://golangci-lint.run/)
    
run tests and linters:

    make test
    make lint
    

## Description

In this assignment, you have to fetch [this list of essays](https://drive.google.com/file/d/1TF4RPuj8iFwpa-lyhxG67V8NDlktmTGi/view) and count the top 10 words from all the essays combined.

A valid word will:
1. Contain at least 3 characters.
2. Contain only alphabetic characters.
3. Be part of our [bank of words](https://raw.githubusercontent.com/dwyl/english-words/master/words.txt) (not all the words in the bank are valid according tothe previous rules)
The output should be pretty JSON printed to the stdout.

## Comments

- Tried to make the project structure as simple as possible. Decided to create a CLI and leave everything inside main package. Separated the logic a bit to make it easier to read and write unit tests.
- All articles are from the same website, so a simple div class selector was implemented for content extraction. Some articles do not exist; such errors are logged and ignored.
- Implemented a worker pool pattern on top of the HTTP requests rate limiting.
- Decided not to separate the article download and parsing logic into different goroutines, as tests showed that the parsing overhead is negligible compared to the article download time.
- The bottleneck here is the request limit from the server side. After receiving too many requests, it blocks the client for approximately 4 minutes and returns 999 status code for any request (even HEAD or OPTIONS). It seems that the server backend has a request limit for a specified time frame. Whether you make a lot of requests and wait for the block, or limit your requests, the long-term result is approximately the same. The limitation can be overcome by using proxy servers, or it is possible to separate the articles list into batches, processing those batches on different machines with further aggregation. This part should be outside the scope of the task. Running the program with 2 RPS limit seems to be safe, but blocks are still possible.
- Added minimal linters and github actions setup.

## Posible improvement

- Proxy usage, horizontal scaling or both (described in the comment section).
- Better error handling with retries. Now, if the website goes down, all failing articles will be skipped.
- Cache statistics from processed articles; if the list changes, it will assist with the next execution.
- Further research on the website's rate limit behaviors is needed to determine the best RPS limits
- The main scenario is covered by unit tests, but coverage is not complete.
- I decided to skip better logging, Docker, versioning, and other similar improvements here.
