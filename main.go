package main

// 6. 现在主流的依赖管理方式是mod，可以参考go的开源代码结构
import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type FileInfo struct {
	Id   int
	Type string
	Size int64
	Name string
}

// 5. 这里在内存里记录文件信息，有没有考虑中间文件更新、修改或者删除等，内存里的内容没同步更新呢？
var fileInfos []FileInfo

// 7. 可以参考下 go 的 web 工程目录结构，都在一个文件里会随着代码复杂度上升，可读性会下降
func main() {
	http.HandleFunc("/download", func(writer http.ResponseWriter, request *http.Request) {
		id := request.FormValue("id")
		atoi, err := strconv.Atoi(id)
		if err != nil {
			writer.Write([]byte("try to parse string to int failed, detail: " + err.Error()))
			return
		}
		// 1. 这里数组如果出界，程序会异常退出
		info := fileInfos[atoi]

		// 2. 可以考虑换一个日志输出库
		fmt.Println(info.Name)
		//err = os.Remove(info.Name)
		//if err != nil {
		//	writer.Write([]byte("remove file failed, detail : " + err.Error()))
		//	return
		//}

		join := filepath.Join("public", info.Name)
		fmt.Println(join)
		file, err := ioutil.ReadFile(join)
		if err != nil {
			writer.Write([]byte("readfile failed, detail : " + err.Error()))
			return
		}

		// 3. 由于文件大小因素，考虑是否压缩？
		writer.Header().Add("Content-Type", "application/octet-stream")
		writer.Header().Add("Content-Disposition", "attachment;filename="+info.Name)
		writer.Write(file)
	})

	http.HandleFunc("/index", func(writer http.ResponseWriter, request *http.Request) {
		dir, err := ioutil.ReadDir("public")
		if err != nil {
			writer.Write([]byte("readDir failed, detail: " + err.Error()))
			return
		}
		fileInfos = []FileInfo{}
		var fileInfo FileInfo
		for i, info := range dir {
			fileInfo.Id = i
			//fileInfo.Type = info.Mode().Type().String()
			fileInfo.Size = info.Size() / 1024
			fileInfo.Name = info.Name()
			fileInfos = append(fileInfos, fileInfo)
		}
		must := template.Must(template.New("index.html").ParseFiles("index.html"))
		fmt.Println(len(fileInfos))
		fmt.Fprint(os.Stdout, len(fileInfos))
		must.Execute(writer, fileInfos)
	})

	http.HandleFunc("/upload", func(writer http.ResponseWriter, request *http.Request) {
		must := template.Must(template.New("upload.html").ParseFiles("upload.html"))
		must.Execute(writer, nil)
	})

	// 4. URL前缀建议驼峰小写开头
	http.HandleFunc("/SaveItem", func(writer http.ResponseWriter, request *http.Request) {
		src, header, err := request.FormFile("file")
		defer src.Close()

		if err != nil {
			writer.Write([]byte("parseFile failed, detail: " + err.Error()))
			return
		}

		name := filepath.Join("public", header.Filename)
		dst, err := os.Create(name)
		defer dst.Close()

		if err != nil {
			writer.Write([]byte("createFile failed, detail : " + err.Error()))
			return
		}

		_, err = io.Copy(dst, src)
		if err != nil {
			writer.Write([]byte("copy file from src failed, deatil : " + err.Error()))
			return
		}

		http.Redirect(writer, request, "/index", http.StatusTemporaryRedirect)
	})

	err := http.ListenAndServe(":9876", nil)
	if err != nil {
		panic(err)
	}
}
