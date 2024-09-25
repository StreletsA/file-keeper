package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const FileKeeperPort string = ":8080"

var (
	FilesMetaDataFileName string = os.Getenv("FILES_META_DATA_FILE_NAME")
	FilesFolderPath       string = os.Getenv("FILES_FOLDER_PATH")

	fileDataRepo FileMetaDataRepository = createInFileFileMetaDataRepository(FilesMetaDataFileName)
	fileKeeper   FileKeeper             = createInFolderFileKeeper(FilesFolderPath, fileDataRepo)
)

type Writer struct {
	data       []byte
	writeIndex int64
}

func (r *Writer) Write(p []byte) (n int, err error) {
	n = copy(r.data, p)
	r.writeIndex += int64(n)
	return
}

func startServer() {
	fileHandler := http.HandlerFunc(fileEndpoint)
	http.Handle("/file", fileHandler)
	http.ListenAndServe(FileKeeperPort, nil)

	fmt.Println("FileKeeper started")
	fmt.Printf("File with files metadata -> %s\n", FilesMetaDataFileName)
	fmt.Printf("Folder with files -> %s\n", FilesFolderPath)
	fmt.Println("----------")
}

func fileEndpoint(w http.ResponseWriter, request *http.Request) {
	if request.Method == "PUT" {
		processSaveFileRequest(w, request)
	} else if request.Method == "GET" {
		processGetFileRequest(w, request)
	} else {
		w.WriteHeader(415)
	}
}

func processSaveFileRequest(w http.ResponseWriter, request *http.Request) {
	err := request.ParseMultipartForm(1024 * 64 << 20) // maxMemory 64*1024MB
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	file, h, err := request.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	fileContent := make([]byte, h.Size)
	wr := Writer{fileContent, 0}
	_, err = io.Copy(&wr, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	originalFileName := h.Filename

	id, err := fileKeeper.Save(OriginalFile{fileContent, originalFileName})
	if err != nil {
		fmt.Printf("Error, file [%s] not saved. Message -> %s\n", originalFileName, err.Error())

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Printf("File [%s] saved with id [%s]\n", originalFileName, id)

	w.WriteHeader(200)
	w.Write([]byte(id))
}

func processGetFileRequest(w http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(415)
		w.Write([]byte("id param required"))
		return
	}

	file, err := fileKeeper.GetById(id)
	if err != nil {
		fmt.Printf("Download error: id [%s]. Message -> %s\n", id, err.Error())

		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment;filename*=UTF-8''%s`, url.PathEscape(file.Name)))
	w.WriteHeader(200)
	w.Write(file.Content)

	fmt.Printf("Download success: File [%s] with id [%s]\n", file.Name, id)
}
