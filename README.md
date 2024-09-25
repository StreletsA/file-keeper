# File Keeper - simple file storage

### Image building
```
docker build --tag file-keeper .
```

### Container starting
```
docker run -p 1234:8080 --name file-keeper file-keeper
```

### Environments (not required)
- FILES_META_DATA_FILE_NAME - path to file with saved files metadata inside container (default: /store/file-data)
- FILES_FOLDER_PATH - path to folder with saved files inside container (default: /store/files)

## API

### Save file
```
PUT /file

Content-Type: multipart/form-data

part name = file
```

#### Response example (saved file id)
```
81649568-70e5-4234-b69d-546acf8677cc
```

### Get file
```
GET /file?id={file_id}
```
