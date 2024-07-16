# Free Upload API
## [中文](https://github.com/superggfun/uploadAPI/blob/main/README_zh-cn.md)

## Project Description

Free Upload API is a free file hosting solution. This is the server part of the system.

## Installation Instructions

- Download the corresponding files from the `release` section on GitHub along with the `config.json`.
- The system is ready to use right out of the box, no further setup is needed.

## Usage

For usage instructions, please refer to the README.md file included in the download.

## File Upload API Documentation

**URL:** /upload

**Method:** POST

**Description:** Uploads files to the server.

### Configuration File

The API's configuration is managed through the `config.json` file, an example is as follows:

```json
{
    "port": "8080",
    "uploadPath": "/upload",
    "maxConcurrentUploads": 100,
    "maxUploadSizeMB": 20,
    "uploadTimeout": 30,
}
```

#### Configuration File Parameters

| Parameter Name       | Type   | Description                          |
| -------------------- | ------ | ------------------------------------ |
| port                 | string | The port on which the server listens |
| uploadPath           | string | The path where files are uploaded    |
| maxConcurrentUploads | int    | Maximum number of concurrent uploads |
| maxUploadSizeMB      | int    | Maximum upload size per form (in MB) |
| uploadTimeout        | int    | Upload timeout (in seconds)          |


### Request Headers

| Key          | Value               |
| ------------ | ------------------- |
| Content-Type | multipart/form-data |

### Request Parameters

The request should include one or more files, uploaded using `multipart/form-data` encoding.

| Parameter Name | Type | Required | Description    |
| -------------- | ---- | -------- | -------------- |
| uploadFile     | file | 是       | File to upload |

### Sample Request

```bash
curl -X POST http://localhost:8080/upload \
  -F "uploadFile=@path/to/your/file1" \
  -F "uploadFile=@path/to/your/file2"
```


### Successful Response

```json
[
    {
        "success": true,
        "code": 200,
        "data": {
            "uploadPath": "http://p0.meituan.net/csc/d1e57fe2aabb918347eb457b081e3f9623243.jpg",
            "filename": "test2.jpg",
            "fileType": "jpg",
            "fileSize": "23243"
        },
        "errMsg": ""
    },
    {
        "success": true,
        "code": 200,
        "data": {
            "uploadPath": "http://p1.meituan.net/csc/7081ac477da112da87b5e8c58913f68c824847.png",
            "filename": "test1.png",
            "fileType": "png",
            "fileSize": "824847"
        },
        "errMsg": ""
    }
]

```

### Failure Response

```json
[
  {
    "success": false,
    "code": 413,
    "data": {
      "uploadPath": "",
      "filename": "",
      "fileType": "",
      "fileSize": ""
    },
    "errMsg": "form is too large. total size needs to be less than specified max upload size"
  }
]
```

### Error Codes

| success | code |                            errMsg                            |
| :-----: | :--: | :----------------------------------------------------------: |
|  true   | 200  |                                                              |
|  false  | 400  | Invalid request, usually due to form parsing failure or file read error |
|  false  | 413  |   Uploaded file size exceeds the configured maximum limit    |
|  false  | 429  | Server is busy, current number of concurrent uploads exceeds the configured limit |
|  false  | 500  |               File does not meet requirements                |
|  false  | 501  | Fragmented upload file parameter validation failed, file information cannot be empty |
|  false  | 502  |                    File validation failed                    |
|  false  |  7   |                        Unknown error                         |

### Supported File Formats

```makefile
.txt, .png, .jpg, .jpeg, .pdf, .docx, .mp3, .mp4, .jpeg, .zip, .apk, .ts, .m3u8
```

### Logs

The server records detailed information about each upload in the `upload.log` file, including client IP, uploaded file name, and upload path.

### Considerations

```makefile
1. The size limit for a single file upload is 20MB.
2. The upload path /upload should be consistent with the file upload path in the server configuration file.
3. Do not upload important files, because they will be published.
```

### Example Code

#### Python

```python
import requests

# List of file paths
file_paths = ["path/to/your/file1.txt", "path/to/your/file2.txt"]
# Upload URL
upload_url = "http://localhost:8080/upload"

# Function to upload files
def upload_files(file_paths):
    for file_path in file_paths:
        with open(file_path, 'rb') as file:
            # Create form data
            files = {'uploadFile': file}

            # Send POST request
            response = requests.post(upload_url, files=files)

            # Print response
            if response.status_code == 200:
                response_data = response.json()
                for result in response_data:
                    if result['success']:
                        print("File uploaded successfully!")
                        print(f"Upload Path: {result['data']['uploadPath']}")
                        print(f"Filename: {result['data']['filename']}")
                        print(f"FileType: {result['data']['fileType']}")
                        print(f"FileSize: {result['data']['fileSize']}")
                    else:
                        print("File upload failed.")
                        print(f"Error Code: {result['code']}")
                        print(f"Error Message: {result['errMsg']}")
            else:
                print(f"Failed to upload file. Status code: {response.status_code}")
                print(f"Response: {response.text}")

# Execute file upload
upload_files(file_paths)
```

### Feedback and Support

If you encounter any issues or have suggestions during use, please contact the development team.
