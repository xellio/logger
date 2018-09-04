## logger

Like ```tail -F /log/this/file```

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
        log.Fatal("No parsable logfiles detected.")
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
