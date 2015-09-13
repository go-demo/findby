# findby

> 以并发的方式查找文件内容

## 获取

``` shell
	$ go get github.com/LyricTian/findby
```

## 范例

> 查找$GOPATH下的所有go文件中的"fmt"

``` shell
$ findby -r "(?i:fmt)" -n $GOPATH -e "go" -o temp.txt
```

## 参数说明

``` shell
	--regexp, -r 				使用正则表达式过滤文件内容
	--name, -n [--name option --name option]	指定过滤的目录名或文件名
	--out, -o 					指定输出文件
	--ext, -e [--ext option --ext option]	指定过滤的文件扩展名
	--count, -c "50"				指定每次并发读取的文件数量
	--help, -h					show help
	--version, -v				print the version
```

