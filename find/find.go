package find

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
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
	// 错误处理
	chError chan error
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

// Find 查找文件内容，并返回查找结果
func (f *File) Find() <-chan FileContent {
	chFileContent := make(chan FileContent, f.ChanCount)
	go f.checkError()
	chFile := f.readFileList()
	go f.findFileContent(chFile, chFileContent)

	return chFileContent
}

func (f *File) checkError() {
	if err := <-f.chError; err != nil {
		fmt.Println("===> Execute error:", err)
		os.Exit(-1)
	}
}

// checkFileExt 检查文件扩展名
func (f *File) checkFileExt(fileName string) bool {
	ext := filepath.Ext(fileName)
	if ext == "" || len(f.Exts) == 0 {
		return ext != ""
	}
	for i := 0; i < len(f.Exts); i++ {
		if string(ext[1:]) == f.Exts[i] {
			return true
		}
	}
	return false
}

// readDir 递归读取目录
func (f *File) readDir(dirName string, chFile chan<- string, count *int64) {
	dir, err := os.Open(dirName)
	if err != nil {
		f.chError <- err
		return
	}
	defer dir.Close()
	fi, err := dir.Readdir(0)
	if err != nil {
		f.chError <- err
		return
	}
	for _, item := range fi {
		name := dirName + "/" + item.Name()
		if item.IsDir() {
			f.readDir(name, chFile, count)
			continue
		}
		if f.checkFileExt(name) {
			*count = *count + 1
			chFile <- name
		}
	}
}

// readFileList 读取文件列表
func (f *File) readFileList() <-chan string {
	chFile := make(chan string, f.ChanCount)
	go func() {
		var fCount int64
		for i := 0; i < len(f.Names); i++ {
			name := f.Names[i]
			if !filepath.IsAbs(name) {
				name, _ = filepath.Abs(name)
			}
			fileInfo, err := os.Stat(name)
			if err != nil {
				f.chError <- err
				break
			}
			if !fileInfo.IsDir() && f.checkFileExt(name) {
				chFile <- name
				continue
			}
			f.readDir(name, chFile, &fCount)
		}
		fmt.Println("++++++> 文件总数量：", fCount)
		close(chFile)
	}()
	return chFile
}

// findFileContent 查找文件内容
func (f *File) findFileContent(chFile <-chan string, chFileContent chan<- FileContent) {
	var (
		mutex    sync.Mutex
		fileSeed int64
		readSeed int64
		chSeed   = make(chan bool, 1)
	)

	go func() {
		if <-chSeed {
			fmt.Println("######> Close file content")
			close(chFileContent)
		}
	}()

	for fItem := range chFile {
		fileSeed++
		go func(name string) {
			file, err := os.Open(name)
			if err != nil {
				f.chError <- err
				return
			}
			defer file.Close()
			reader := bufio.NewReader(file)
			var (
				fc      FileContent
				buffer  = new(bytes.Buffer)
				lineNum = 1
			)
			fc.FileName = name
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
			mutex.Lock()
			readSeed++
			mutex.Unlock()
			fmt.Println("######> 查找文件数量：", readSeed)
			if readSeed == fileSeed {
				chSeed <- true
			}
		}(fItem)
	}
	fmt.Println("------> 读取文件数量：", fileSeed)
}
