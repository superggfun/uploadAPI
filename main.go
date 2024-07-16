// Copyright (c) 2024 superggfun
// All rights reserved.

////////////////////////////////////////////////////////////////////
//                          _ooOoo_                               //
//                         o8888888o                              //
//                         88" . "88                              //
//                         (| ^_^ |)                              //
//                         O\  =  /O                              //
//                      ____/`---'\____                           //
//                    .'  \\|     |//  `.                         //
//                   /  \\|||  :  |||//  \                        //
//                  /  _||||| -:- |||||-  \                       //
//                  |   | \\\  -  /// |   |                       //
//                  | \_|  ''\---/''  |   |                       //
//                  \  .-\__  `-`  ___/-. /                       //
//                ___`. .'  /--.--\  `. . ___                     //
//              ."" '<  `.___\_<|>_/___.'  >'"".                  //
//            | | :  `- \`.;`\ _ /`;.`/ - ` : | |                 //
//            \  \ `-.   \_ __\ /__ _/   .-` /  /                 //
//      ========`-.____`-.___\_____/___.-`____.-'========         //
//                           `=---='                              //
//      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^        //
//         佛祖保佑       永无BUG     永不修改                      //
////////////////////////////////////////////////////////////////////

// This source code is licensed under the MIT-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

// Config 结构体用于解析配置文件
type Config struct {
	Port                 string `json:"port"`
	UploadPath           string `json:"uploadPath"`
	MaxConcurrentUploads int    `json:"maxConcurrentUploads"`
	MaxUploadSizeMB      int    `json:"maxUploadSizeMB"`
	UploadTimeout        int    `json:"uploadTimeout"` // 秒
}

// FileUploader 接口定义上传文件的方法
type FileUploader interface {
	Upload(key string) (*UploadResponse, error)
}

// file 结构体用于处理文件数据和名称，并实现 FileUploader 接口
type file struct {
	fileData []byte
	name     string
}

// NewFileFromPath 从文件路径创建一个 file 对象
func NewFileFromPath(filePath string) *file {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	_, name := filepath.Split(filePath)
	return &file{
		fileData: data,
		name:     name,
	}
}

// NewFileFromBytes 从字节切片和文件名创建一个 file 对象
func NewFileFromBytes(fileData []byte, filename string) *file {
	return &file{
		fileData: fileData,
		name:     filename,
	}
}

var (
	configKey string
	keyMutex  sync.Mutex
)

// getVisitId模拟从服务中获取新的token及其到期时间
func getVisitId() (string, time.Time) {
	// 此处为模拟值，实际应从API响应中获取
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	decodedURL, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9rZi5kaWFucGluZy5jb20vY3NDZW50ZXIvYWNjZXNzL2RlYWxPcmRlcl9IZWxwX0RQX1BD")
	req, err := http.NewRequest(http.MethodGet, string(decodedURL), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	// 添加请求头
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 确保响应中包含Location头
	location, ok := resp.Header["Location"]
	if !ok || len(location) == 0 {
		log.Fatalf("No Location header in response")
	}

	strs := regexp.MustCompile(`visitId=(.*?)&subSource`)
	matches := strs.FindAllStringSubmatch(location[0], -1)
	if len(matches) == 0 || len(matches[0]) < 2 {
		log.Fatalf("No visitId found in Location header")
	}
	token := matches[0][1]

	postBody := `{"type":"Init","parameters":{"isPreview":true,"build":null,"fe":"portal_pc"}}`
	decodedURLStart, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9rZi5kaWFucGluZy5jb20vYXBpL3BvcnRhbC9tZXNzYWdlL2luaXQ/dmlzaXRJZD0=")
	decodedURLEnd, _ := base64.StdEncoding.DecodeString("JmFjY2Vzc1Rva2VuPXVuZGVmaW5lZA==")

	url := string(decodedURLStart) + token + string(decodedURLEnd)

	postReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(postBody)))
	if err != nil {
		log.Fatalf("Error creating POST request: %v", err)
	}
	postReq.Header.Add("Content-Type", "application/json")
	postReq.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	_, err = client.Do(postReq)
	if err != nil {
		log.Fatalf("Error sending POST request: %v", err)
	}
	expiresAt := time.Now().Add(12 * time.Hour) // 假设token每12小时到期
	return token, expiresAt
}

// refreshToken负责调用getVisitId并更新configKey
func refreshToken() {
	keyMutex.Lock()
	defer keyMutex.Unlock()
	newKey, _ := getVisitId()
	configKey = newKey
	fmt.Println("Token refreshed at:", time.Now().Format(time.RFC1123), "New Key:", configKey)
}

// scheduleTokenRefresh 设置定时刷新任务
func scheduleTokenRefresh() {
	refreshToken() // 初始加载
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for {
			<-ticker.C
			refreshToken()
		}
	}()
}

// UploadResponse 结构体用于解析上传文件后的响应
type UploadResponse struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
	Data    struct {
		UploadPath string `json:"uploadPath"`
		Filename   string `json:"filename"`
		FileType   string `json:"fileType"`
		FileSize   string `json:"fileSize"`
	} `json:"data"`
	ErrMsg string `json:"errMsg"`
}

// Upload 上传单个文件到指定的 API
func (f *file) Upload(client *http.Client) (*UploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", f.name)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = part.Write(f.fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	currentTimestamp := time.Now().UnixMilli()

	_ = writer.WriteField("fileName", f.name)
	_ = writer.WriteField("fileID", fmt.Sprintf("%d", currentTimestamp))
	_ = writer.WriteField("part", "0")
	_ = writer.WriteField("partSize", "1")

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9rZi5kaWFucGluZy5jb20vYXBpL2ZpbGUvYnVyc3RVcGxvYWRGaWxl")
	req, err := http.NewRequest("POST", string(url), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	referer, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly9oNS5kaWFucGluZy5jb20v")
	req.Header = http.Header{
		"Content-Type": {writer.FormDataContentType()},
		"User-Agent":   {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"},
		"Accept":       {"application/json, text/plain, */*"},
		"csc-visitid":  {configKey},
		"Referer":      {string(referer)},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server upload request timed out")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var uploadResponse UploadResponse
	err = json.Unmarshal(respBody, &uploadResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &uploadResponse, nil
}

func newHTTPClient(timeout int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}

// uploadHandler 处理上传请求的HTTP处理函数
func uploadHandler(w http.ResponseWriter, r *http.Request, config *Config, logger *log.Logger) {
	if r.Method != http.MethodPost {
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusMethodNotAllowed,
			ErrMsg:  "invalid request method",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}

	maxUploadSize := int64(config.MaxUploadSizeMB) << 20 // 配置中的表单最大大小
	const maxSingleFileSize = 20 << 20

	if r.ContentLength > maxUploadSize {
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusRequestEntityTooLarge,
			ErrMsg:  "form is too large. total size needs to be less than specified max upload size",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}

	err := r.ParseMultipartForm(maxUploadSize) // 配置中的最大文件大小
	if err != nil {
		logger.Printf("error parsing multipart form: %v", err)
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			ErrMsg:  "unable to parse form",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}

	if r.MultipartForm == nil {
		logger.Println("MultipartForm is nil")
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusInternalServerError,
			ErrMsg:  "multipart form is nil",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}

	files := r.MultipartForm.File["uploadFile"]
	if files == nil {
		logger.Println("No files found in the form")
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			ErrMsg:  "no files found in the form",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}

	requestSem := semaphore.NewWeighted(int64(config.MaxConcurrentUploads))
	if err := requestSem.Acquire(r.Context(), 1); err != nil {
		response := &UploadResponse{
			Success: false,
			Code:    http.StatusTooManyRequests,
			ErrMsg:  "server too busy",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*UploadResponse{response})
		return
	}
	defer requestSem.Release(1)

	client := newHTTPClient(config.UploadTimeout)

	var wg sync.WaitGroup
	fileSem := semaphore.NewWeighted(int64(config.MaxConcurrentUploads))
	responseChan := make(chan *UploadResponse, len(files))

	for _, handler := range files {
		if handler == nil {
			logger.Println("File handler is nil")
			continue
		}

		wg.Add(1)

		go func(handler *multipart.FileHeader) {
			defer wg.Done()

			if err := fileSem.Acquire(r.Context(), 1); err != nil {
				responseChan <- &UploadResponse{
					Success: false,
					Code:    http.StatusTooManyRequests,
					ErrMsg:  "server too busy",
				}
				return
			}
			defer fileSem.Release(1)

			file, err := handler.Open()
			if err != nil {
				responseChan <- &UploadResponse{
					Success: false,
					Code:    http.StatusBadRequest,
					ErrMsg:  "error retrieving the file",
				}
				return
			}
			defer file.Close()

			if handler.Size > maxSingleFileSize {
				responseChan <- &UploadResponse{
					Success: false,
					Code:    http.StatusRequestEntityTooLarge,
					ErrMsg:  "file is too large. file size needs to be less than 20MB",
				}
				return
			}

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				responseChan <- &UploadResponse{
					Success: false,
					Code:    http.StatusInternalServerError,
					ErrMsg:  "error reading the file",
				}
				return
			}
			f := NewFileFromBytes(fileBytes, handler.Filename)
			uploadResponse, err := f.Upload(client)
			if err != nil || !uploadResponse.Success {
				if uploadResponse == nil {
					uploadResponse = &UploadResponse{}
				}
				uploadResponse.Success = false
				uploadResponse.Code = http.StatusInternalServerError
				uploadResponse.ErrMsg = fmt.Sprintf("file upload failed: %v", err)
			}
			clientIP := r.RemoteAddr
			logEntry := fmt.Sprintf("IP: %s, Code: %d, Filename: %s, UploadPath: %s\n",
				clientIP, uploadResponse.Code, handler.Filename, uploadResponse.Data.UploadPath)
			logger.Println(logEntry)

			responseChan <- uploadResponse
		}(handler)
	}

	wg.Wait()
	close(responseChan)

	var responses []*UploadResponse
	for res := range responseChan {
		responses = append(responses, res)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func loadConfig(configFile string) (*Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logFile, err := os.OpenFile("upload.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	scheduleTokenRefresh() // 启动定时刷新任务
	http.HandleFunc(config.UploadPath, func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(w, r, config, logger)
	})

	log.Printf("server starting on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
