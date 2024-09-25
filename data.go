package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultFilesMetaDataFileName string = "/store/file-data"
	DefaultFilesFolderName       string = "/store/files"
)

type FileMetaData struct {
	id           string
	originalName string
	creatingDate int
}

type FileMetaDataRepository interface {
	Save(fileMetaData FileMetaData) (string, error)
	GetById(id string) (FileMetaData, error)
}

type InFileFileMetaDataRepository struct {
	StorageFilePath string
}

func createInFileFileMetaDataRepository(storageFilePath string) InFileFileMetaDataRepository {
	if storageFilePath == "" {
		fmt.Printf("Files metadata file name empty. Default file name using -> %s\n", DefaultFilesMetaDataFileName)
		storageFilePath = DefaultFilesMetaDataFileName
	}
	return InFileFileMetaDataRepository{storageFilePath}
}

func (repo InFileFileMetaDataRepository) Save(fileMetaData FileMetaData) (string, error) {
	if repo.StorageFilePath == "" {
		return "", errors.New("StorageFilePath is empty")
	}

	file, err := os.OpenFile(repo.StorageFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("opening file-data file error: %s", err.Error())
		if errors.Is(err, fs.ErrNotExist) {
			file, err = os.Create(repo.StorageFilePath)
		}

		if err != nil {
			fmt.Printf("creating file-data file error: %s\n", err.Error())
			return "", err
		}
	}

	defer file.Close()

	_, err = file.Write([]byte(fmt.Sprintf("%s,%s,%d\n", fileMetaData.id, fileMetaData.originalName, fileMetaData.creatingDate)))
	if err != nil {
		fmt.Printf("writing file-data file error: %s\n", err.Error())
		return "", err
	}

	return fileMetaData.id, nil
}

func (repo InFileFileMetaDataRepository) GetById(id string) (FileMetaData, error) {
	if repo.StorageFilePath == "" {
		return FileMetaData{}, errors.New("StorageFilePath is empty")
	}

	content, err := os.ReadFile(repo.StorageFilePath)
	if err != nil {
		fmt.Printf("reading file-data file: %s\n", err.Error())
		return FileMetaData{}, err
	}

	storage := string(content)
	lines := strings.Split(storage, "\n")

	for _, line := range lines {
		record := strings.Split(line, ",")
		if len(record) < 3 {
			continue
		}

		recordId := record[0]

		if recordId == id {
			originalName := record[1]
			creatingDate, err := strconv.Atoi(record[2])
			if err != nil {
				creatingDate = 0
			}

			return FileMetaData{id, originalName, creatingDate}, nil
		}
	}

	return FileMetaData{}, errors.New("file not found")
}

type OriginalFile struct {
	Content []byte
	Name    string
}

type FileKeeper interface {
	Save(file OriginalFile) (string, error)
	GetById(id string) (OriginalFile, error)
}

type InFolderFileKeeper struct {
	FolderName             string
	FileMetaDataRepository FileMetaDataRepository
}

func createInFolderFileKeeper(folderName string, fileMetaDataRepository FileMetaDataRepository) InFolderFileKeeper {
	if folderName == "" {
		fmt.Printf("Files folder name empty. Default folder name using -> %s\n", DefaultFilesFolderName)
		folderName = DefaultFilesFolderName
	}

	if fileMetaDataRepository == nil {
		panic("FileMetaDataRepository is nil")
	}

	return InFolderFileKeeper{folderName, fileMetaDataRepository}
}

func (fileKeeper InFolderFileKeeper) Save(file OriginalFile) (string, error) {
	if fileKeeper.FolderName == "" {
		return "", errors.New("FolderName is empty\n")
	}

	id := uuid.New().String()

	err := os.MkdirAll(fileKeeper.FolderName, os.ModePerm)
	if err != nil {
		return "", err
	}

	f, err := os.Create(fmt.Sprintf("%s/%s", fileKeeper.FolderName, id))
	if err != nil {
		fmt.Printf("creating file error: %s\n", err.Error())
		return "", err
	}

	defer f.Close()

	_, err = f.Write(file.Content)
	if err != nil {
		fmt.Printf("writing file error: %s\n", err.Error())
		return "", err
	}

	savedId, err := fileKeeper.FileMetaDataRepository.Save(FileMetaData{id, file.Name, int(time.Now().Unix())})
	if err != nil {
		fmt.Printf("saving file error: %s\n", err.Error())
		return "", err
	}

	return savedId, nil
}

func (fileKeeper InFolderFileKeeper) GetById(id string) (OriginalFile, error) {
	if fileKeeper.FolderName == "" {
		return OriginalFile{}, errors.New("FolderName is empty")
	}

	fileMetaData, err := fileKeeper.FileMetaDataRepository.GetById(id)
	if err != nil {
		return OriginalFile{}, err
	}

	content, err := os.ReadFile(fmt.Sprintf("%s/%s", fileKeeper.FolderName, fileMetaData.id))
	if err != nil {
		return OriginalFile{}, err
	}

	return OriginalFile{content, fileMetaData.originalName}, nil
}
