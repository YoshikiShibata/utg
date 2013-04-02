// Copyright (C) 2013 Yoshiki Shibata. All rights reserved

package main

import "fmt"
import "os"
import "bufio"
import "regexp"

// This is a simple "grep" command implementation
// Each specified file will be examined by a goroutine assigned for the file.

// grep [OPTIONS] PATTERN FILE...
//
// Note that the current implementation supports NO OPTIONS.

func main() {
    args := os.Args
    if len(args) <= 2 {
        showUsage(args[0])
        os.Exit(1)
    }

    pattern := args[1];
    files := args[2:]

    grepPatternFromFiles(pattern, files)
    os.Exit(0)
}

func showUsage(programName string) {
    fmt.Printf("Version 0.0\n")
    fmt.Printf("%s [OPTIONS] PATTERN FILE...\n", programName)
}

type grepResult struct {
    file        string
    eof         bool
    lineNumber  int
    line        string
}


func grepPatternFromFiles(pattern string, files []string) {
    compiledPattern, err := regexp.Compile(pattern)
    if err != nil {
        fmt.Printf("Illegal pattern (%s) : %s\n", pattern, err.Error())
        os.Exit(1)
    }

    results := make([]chan grepResult, len(files))
    for i := 0; i < len(files); i++ {
        results[i] = make(chan grepResult)
        go grepPatternFromOneFile(compiledPattern, files[i], results[i])
    }

    showResults(results)
}

func showResults(results []chan grepResult) {
    for i := 0; i < len(results); i++ {
        for result := <- results[i];
            !result.eof;
            result = <- results[i] {
            fmt.Printf("%s(%d): %s\n",
                    result.file,
                    result.lineNumber,
                    result.line)
        }
    }
}

func grepPatternFromOneFile(pattern *regexp.Regexp,
                 file string,
                 resultChan chan grepResult) {
    var result grepResult

    result.file = file
    result.eof  = false

    openedFile, err := os.Open(file)
    if err != nil {
        fmt.Printf("%s: %s\n", file, err.Error())
        result.eof = true
        resultChan <- result
        return
    }

    lineNumber := 0
    lineReader := bufio.NewReaderSize(openedFile, 255)
    for line, isPrefix, e := lineReader.ReadLine()
        e == nil
        line, isPrefix, e = lineReader.ReadLine() {
        lineNumber++
        fullLine := string(line)
        if isPrefix {
            for {
                line, isPrefix, _ = lineReader.ReadLine()
                fullLine = fullLine + string(line)
                if !isPrefix {
                    break
                }
            }
        }
        if pattern.MatchString(fullLine) {
            result.lineNumber = lineNumber
            result.line = fullLine
            resultChan <- result
        }
    }

    result.eof = true
    resultChan <- result
}
