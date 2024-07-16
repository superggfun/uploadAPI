## 文件上传API接口文档

**URL:** /upload

**方法:** POST

**描述:** 上传文件到服务器。

### 配置文件

API的配置通过 `config.json` 文件进行管理，示例如下：

```json
{
    "port": "8080",
    "uploadPath": "/upload",
    "maxConcurrentUploads": 100,
    "maxUploadSizeMB": 20,
    "uploadTimeout": 30,
}
```

#### 配置文件参数
| 参数名               | 类型   | 描述                             |
| -------------------- | ------ | -------------------------------- |
| port                 | string | 服务器监听的端口                 |
| uploadPath           | string | 文件上传的路径                   |
| maxConcurrentUploads | int    | 最大并发上传数                   |
| maxUploadSizeMB      | int    | 表单的最大上传大小（以MB为单位） |
| uploadTimeout        | int    | 上传超时时间（以秒为单位）       |


### 请求头

| Key          | Value               |
| ------------ | ------------------- |
| Content-Type | multipart/form-data |

### 请求参数

请求应包含一个或多个文件，使用 `multipart/form-data` 编码方式上传。

| 参数名     | 类型 | 必填 | 描述         |
| ---------- | ---- | ---- | ------------ |
| uploadFile | file | 是   | 要上传的文件 |

### 示例请求
```bash
curl -X POST http://localhost:8080/upload \
  -F "uploadFile=@path/to/your/file1" \
  -F "uploadFile=@path/to/your/file2"
```


### 成功响应

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

### 失败响应

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

### 错误码

| success | code |                      errMsg                      |
| :-----: | :--: | :----------------------------------------------: |
|  true   | 200  |                                                  |
|  false  | 400  | 请求无效，通常由于表单解析失败或文件读取错误引起 |
|  false  | 413  |        上传的文件大小超过了配置的最大限制        |
| falses  | 429  | 服务器繁忙，当前并发上传数量超过了配置的最大限制 |
|  false  | 500  |                  文件不符合要求                  |
|  false  | 501  |    分片上传文件参数校验失败，文件信息不能为空    |
|  false  | 502  |                   文件校验失败                   |
|  false  |  7   |                     未知错误                     |

### 支持的文件格式

```makefile
.txt, .png, .jpg, .jpeg, .pdf, .docx, .mp3, .mp4, .jpeg, .zip, .apk, .ts, .m3u8
```

### 日志

服务器会在 `upload.log` 文件中记录每次上传的详细信息，包括客户端IP、上传文件名和上传路径。

### 注意事项

```makefile
1. 上传单个文件的大小限制为 20MB。
2. 上传路径 /upload 应与服务器配置文件中的文件上传的路径一致。
```

### 示例代码

#### Python

```python
import requests

# 文件路径列表
file_paths = ["path/to/your/file1.txt", "path/to/your/file2.txt"]
# 上传地址
upload_url = "http://localhost:8080/upload"

# 上传文件函数
def upload_files(file_paths):
    for file_path in file_paths:
        with open(file_path, 'rb') as file:
            # 创建表单数据
            files = {'uploadFile': file}

            # 发送POST请求
            response = requests.post(upload_url, files=files)

            # 打印响应
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

# 执行文件上传
upload_files(file_paths)

```
### 反馈与支持
如果您在使用过程中遇到问题或有任何建议，请联系开发团队