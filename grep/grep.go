// Copyright (C) 2013 Yoshiki Shibata. All rights reserved

package main

import "fmt"
import "os"
import "bufio"
import "regexp"
import "io/ioutil"
import "runtime"

// This is a simple "grep" command implementation
// Each specified file will be examined by a goroutine assigned for the file.

// grep [-r] PATTERN FILE...
//
// Note that the current implementation supports NO OPTIONS.

func main() {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

    args := os.Args
    if len(args) <= 2 {
        showUsage(args[0])
        os.Exit(1)
    }

    pattern := args[1];
    files := args[2:]

	grep := newGoGrep(pattern, files)
	reduceChan := make(chan grepResult)
    go grep.executeFromDirectory(reduceChan, ".")
	showResults(reduceChan)
    os.Exit(0)
}

type goGrep struct {
	pattern 		string
	compiledPattern *regexp.Regexp
	files[] 		string
}

func newGoGrep(pattern string, files []string) *goGrep {
	this := new(goGrep)
	this.pattern = pattern
	this.files = files

	compiledPattern, err := regexp.Compile(pattern)
    if err != nil {
        fmt.Printf("Illegal pattern (%s) : %s\n", pattern, err.Error())
        os.Exit(1)
    }
	this.compiledPattern = compiledPattern
	return this
}

func showUsage(programName string) {
    fmt.Printf("Version 0.2\n")
    fmt.Printf("%s [-r] PATTERN FILE...\n", programName)
}

type grepResult struct {
    file        string
    lineNumber  int
    line        string
}


func (this *goGrep) executeFromDirectory(result    chan grepResult, 
										 directory string) {
    expandedFiles := expandFiles(directory, this.files)

	noOfFiles := len(expandedFiles)
    results := make([]chan grepResult, noOfFiles)

	for i := 0; i < noOfFiles; i++ {
        results[i] = make(chan grepResult)
	}

    go reduceResults(result, results)

    for i, file := range expandedFiles {
		fileInfo, err := os.Stat(file)
    	if err != nil {
        	fmt.Printf("Illegal file (%s) : %s\n", file, err.Error())
        	os.Exit(1)
		}
		if fileInfo.IsDir() {
			go this.executeFromDirectory(results[i], file)
		} else {
        	go this.grepPatternFromOneFile(file, results[i])
		}
    }
}

func expandFiles(directory string, files []string) []string {
    result := make([]string, 0, len(files))

    for _, file := range files {
        for _, expandedFile := range expandFile(directory, file) {
            result = append(result, expandedFile)
        }
    }
    return result
}

func expandFile(directory string, file string) []string {
    result := make([]string, 0, 1)

	pattern := toRegexPattern(file)
	compiledPattern, err := regexp.Compile(pattern)
    if err != nil {
        fmt.Printf("Illegal pattern (%s) : %s\n", pattern, err.Error())
        os.Exit(1)
    }

	fileInfos, _ := ioutil.ReadDir(directory)
	for _, fileInfo := range fileInfos {
		fileName := fileInfo.Name()
		if compiledPattern.MatchString(fileName) || 
		   fileInfo.IsDir() {
			if (directory == ".") {
				result = append(result, fileName)
			} else {
				result = append(result, directory + "/" + fileName)
			}
		}
	}

	return result
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

func reduceResults(reduceChan chan grepResult, results []chan grepResult) {
    for _, resultChan := range results {
        for result, ok := <- resultChan;
            ok;
            result, ok = <- resultChan {
			reduceChan <- result
        }
    }
	close(reduceChan)
}

func showResults(resultsChan chan grepResult) {
	for result, ok := <- resultsChan;
	    ok;
		result, ok = <- resultsChan {
        fmt.Printf("%s(%d): %s\n",
                    result.file,
                    result.lineNumber,
                    result.line)
	}
}


func (this *goGrep) grepPatternFromOneFile(file string,
                 resultChan chan grepResult) {
    var result grepResult

    result.file = file

    openedFile, err := os.Open(file)
    if err != nil {
        fmt.Printf("%s: %s\n", file, err.Error())
		close(resultChan)
        return
    }

    defer openedFile.Close()

	// runtime.Gosched()

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
        if this.compiledPattern.MatchString(fullLine) {
            result.lineNumber = lineNumber
            result.line = fullLine
            resultChan <- result
        }
    }

	close(resultChan)
}
