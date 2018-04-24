package main

import (
	"os"
	"fmt"
	"io"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}


/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}


func main(){

	var wireteString = "[09/Jan/2006:15:04:05 +0000] http \"GET /foo?query=t HTTP/1.0\" 200 2133 \"-\" \"KeepAliveClient\" \"-\" 1.005 1.854\n"
	var filename = "./access.log"
	var f *os.File
	var err1 error

	if checkFileIsExist(filename) { //如果文件存在
		f, err1 = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
		fmt.Println("文件存在")
	} else {
		f, err1 = os.Create(filename) //创建文件
		fmt.Println("文件不存在")
	}
	check(err1)
	for true {
		io.WriteString(f, wireteString) //写入文件(字符串)
		time.Sleep(1*time.Second)
	}

}