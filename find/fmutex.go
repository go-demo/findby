package find

import "sync"

// FMutex 监控过滤文件
type FMutex struct {
	mutex       *sync.Mutex
	readCount   int64
	findCount   int64
	resultCount int64
	chRead      chan bool
}

// ReadCount 获取读取文件数量
func (f *FMutex) ReadCount() int64 {
	return f.readCount
}

// ResultCount 获取查找结果数量
func (f *FMutex) ResultCount() int64 {
	return f.resultCount
}
