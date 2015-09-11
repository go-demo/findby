package find

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"fmt"
)

// NewFile 创建File指针
func NewFile(names, exts []string, reg string, count int) *File {
	r, err := regexp.Compile(reg)
	if err != nil {
		panic(err)
	}
	return &File{
		Names:     names,
		Exts:      exts,
		Regexp:    r,
		ChanCount: count,
		mutex:     fMutex{mutex: new(sync.Mutex), chRead: make(chan bool, 1)},
	}
}

// File 查找文件内容
type File struct {
	// 查找的目录或文件名
	Names []string
	// 查找的文件扩展名
	Exts []string
	// 查找内容的正则表达式
	Regexp *regexp.Regexp
	// 缓冲区的文件数量
	ChanCount int
	// 记录文件查找数量
	mutex fMutex
}

// FileContent 文件内容
type FileContent struct {
	// 文件名
	FileName string
	// 查找的行内容
	Lines []Line
}

// Line 查找的行
type Line struct {
	// 行号
	Number int
	// 行内容
	Content string
}

type fMutex struct {
	mutex     *sync.Mutex
	readCount int64
	findCount int64
	chRead    chan bool
}

// Find 查找文件内容，并返回查找结果
func (f *File) Find() <-chan FileContent {
	var (
		chFileContent = make(chan FileContent, f.ChanCount)
		chFilePath    = make(chan string, f.ChanCount)
	)

	go f.readFileList(chFilePath)

	go f.findContent(chFilePath, chFileContent)

	return chFileContent
}

func (f *File) handleError(err interface{}) {
	fmt.Println("===> Perform error:", err)
	os.Exit(-1)
}

// checkFileExt 检查文件扩展名
func (f *File) checkFileExt(fileName string) bool {
	ext := filepath.Ext(fileName)
	if ext == "" || len(f.Exts) == 0 {
		return ext != ""
	}
	for i := 0; i < len(f.Exts); i++ {
		if string(ext[1:]) == strings.Trim(f.Exts[i], " ") {
			return true
		}
	}
	return false
}

// readDir 递归读取目录
func (f *File) readDir(dirName string, chFilePath chan<- string) {
	defer func() {
		if err := recover(); err != nil {
			f.handleError(err)
		}
	}()
	dir, err := os.Open(dirName)
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	fi, err := dir.Readdir(0)
	if err != nil {
		panic(err)
	}
	for _, item := range fi {
		name := dirName + "/" + item.Name()
		if item.IsDir() {
			f.readDir(name, chFilePath)
			continue
		}
		if f.checkFileExt(name) {
			f.mutex.readCount++
			chFilePath <- name
		}
	}
}

// readFileList 读取文件列表
func (f *File) readFileList(chFilePath chan<- string) {
	defer func() {
		if err := recover(); err != nil {
			f.handleError(err)
		}
		f.mutex.chRead <- true
		close(chFilePath)
	}()
	for i := 0; i < len(f.Names); i++ {
		var (
			err  error
			name = f.Names[i]
		)
		if !filepath.IsAbs(name) {
			name, err = filepath.Abs(name)
			if err != nil {
				panic(err)
			}
		}
		fileInfo, err := os.Stat(name)
		if err != nil {
			panic(err)
		}
		if !fileInfo.IsDir() && f.checkFileExt(name) {
			f.mutex.readCount++
			chFilePath <- name
			continue
		}
		f.readDir(name, chFilePath)
	}
}

func (f *File) findContent(chFilePath <-chan string, chFileContent chan<- FileContent) {
	for filePath := range chFilePath {
		go f.findItem(filePath, chFileContent)
	}
}

func (f *File) findItem(filePath string, chFileContent chan<- FileContent) {
	defer func() {
		if err := recover(); err != nil {
			f.handleError(err)
		}
		f.mutex.mutex.Unlock()
	}()
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var (
		fc      = FileContent{FileName: filePath}
		buffer  = new(bytes.Buffer)
		lineNum = 1
	)
	reader := bufio.NewReader(file)

	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		if isPrefix {
			buffer.Write(line)
			continue
		}
		var l []byte
		if buffer.Len() > 0 {
			buffer.Write(line)
			l = buffer.Bytes()
			buffer.Reset()
		} else {
			l = line
		}
		if f.Regexp.Match(l) {
			fc.Lines = append(fc.Lines, Line{Number: lineNum, Content: string(l)})
		}
		lineNum++
	}
	if len(fc.Lines) > 0 {
		chFileContent <- fc
	}
	f.mutex.mutex.Lock()
	f.mutex.findCount++
	select {
	case <-f.mutex.chRead:
		if f.mutex.readCount == f.mutex.findCount {
			close(chFileContent)
		} else {
			f.mutex.chRead <- true
		}
	default:
	}
}
