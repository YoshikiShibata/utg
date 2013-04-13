// Copyright (C) 2013 Yoshiki Shibata. All rights reserved

package main

import "fmt"
import "os"
import "bufio"
import "regexp"
import "io/ioutil"

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
    lineNumber  int
    line        string
}


func grepPatternFromFiles(pattern string, files []string) {
    expandedFiles := expandFiles(files)
    compiledPattern, err := regexp.Compile(pattern)
    if err != nil {
        fmt.Printf("Illegal pattern (%s) : %s\n", pattern, err.Error())
        os.Exit(1)
    }

    results := make([]chan grepResult, len(expandedFiles))

    for i, file := range expandedFiles {
        results[i] = make(chan grepResult)
        go grepPatternFromOneFile(compiledPattern, file, results[i])
    }

    showResults(results)
}

func expandFiles(files []string) []string {
    result := make([]string, 0, len(files))

    for _, file := range files {
        for _, expandedFile := range expandFile(file) {
            result = append(result, expandedFile)
        }
    }
    return result
}

// NOT IMPLEMENTED YET
func expandFile(file string) []string {
    result := make([]string, 0, 1)

	if !containsAsterisks(file) {
    	return append(result, file);
	}

	pattern := toRegexPattern(file)
	compiledPattern, err := regexp.Compile(pattern)
    if err != nil {
        fmt.Printf("Illegal pattern (%s) : %s\n", pattern, err.Error())
        os.Exit(1)
    }


	fileInfos, _ := ioutil.ReadDir(".")
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if compiledPattern.MatchString(fileName) {
			result = append(result, fileName)
		}
	}

	return result
}

func containsAsterisks(fileName string) bool {
	for _, ch := range fileName {
		if ch == '*' {
			return true
		}
	}
	return false
}

func toRegexPattern(fileName string) string {
	runes := make([]rune, 0, len(fileName))
	for _, ch := range fileName {
		if ch == '*' {
			runes = append(runes,'.')
			runes = append(runes,'*')
		} else {
			runes = append(runes, ch)
		}
	}
	return string(runes)
}

func showResults(results []chan grepResult) {
    for _, resultChan := range results {
        for result, ok := <- resultChan;
            ok;
            result, ok = <- resultChan {
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

    openedFile, err := os.Open(file)
    if err != nil {
        fmt.Printf("%s: %s\n", file, err.Error())
		close(resultChan)
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

	close(resultChan)
}
