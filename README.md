## logger

Like ```tail -F /log/this/file```

[![go report card](https://goreportcard.com/badge/github.com/xellio/logger "go report card")](https://goreportcard.com/report/github.com/xellio/logger)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/xellio/logger?status.svg)](https://godoc.org/github.com/xellio/logger)

Usage:
```
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/xellio/logger"
)

var logFiles []string

func init() {
    for _, file := range os.Args[1:] {
        file = filepath.Clean(file)
        logFiles = append(logFiles, file)
    }
}

func main() {

    if len(logFiles) <= 0 {
        fmt.Println(`Usage:
    logger /path/to/logfile /path/to/another/logfile`)
        os.Exit(1)
    }

    resCh := make(chan *logger.Change)
    errCh := make(chan error)
    go func() {
        err := logger.Start(resCh, logFiles...)
        if err != nil {
            errCh <- err
        }
    }()

    for {
        select {
        case change := <-resCh:
            for _, line := range change.Lines {
                fmt.Println(string(line.Content))
            }

        case err := <-errCh:
            log.Fatalf("Error: %s", err.Error())
            os.Exit(1)
        }
    }

}

```
