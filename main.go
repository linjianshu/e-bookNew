package main

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

var fileInfos []FileInfo

func main() {
	http.HandleFunc("/download", func(writer http.ResponseWriter, request *http.Request) {
		id := request.FormValue("id")
		atoi, err := strconv.Atoi(id)
		if err != nil {
			writer.Write([]byte("try to parse string to int failed, detail: " + err.Error()))
			return
		}
		info := fileInfos[atoi]
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

	//http.Handle("/static", http.StripPrefix("/static/", http.FileServer(http.Dir("public"))))

	err := http.ListenAndServe(":9876", nil)
	if err != nil {
		panic(err)
	}
}
