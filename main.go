package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type ImageUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

type Config struct {
	ApiUrl string `json:"ApiUrl"`
	Token  string `json:"token"`
}

func loadConfig() (*Config, error) {
	// 获取当前运行目录
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
	}
	executableDir := filepath.Dir(executablePath)
	fmt.Println("Executable Path:", executablePath) 
	// 拼接配置文件路径
	configFilePath := filepath.Join(executableDir, "config.yaml")
	// 从配置文件中读取配置使用VIP库
	v := viper.New()
	v.SetConfigFile(configFilePath) // 设置配置文件的路径和名称

	// 可以根据需要设置其他配置选项
	v.SetConfigType("yaml") // 指定配置文件的类型为YAML格式

	// 加载配置文件
	if err := v.ReadInConfig(); err != nil {
		panic(err) // 处理配置文件读取错误
	}
	apiurl := v.GetString("client.ApiUrl")
	token := v.GetString("client.Token")
	// 加载配置文件，定义各项参数
	return &Config{
		apiurl,
		token,
	}, nil
}

func main() {
	// 读取配置文件
	// 读取配置文件
	config, err := loadConfig()
	if err != nil {
		return
	}
	// 准备上传数据
	imagePath := os.Args[len(os.Args)-1]
	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file '%v': %v\n", imagePath, err)
		os.Exit(1)
	}

	// 接口URL
	url := config.ApiUrl

	// 发送上传请求
	resBody, err := uploadImage(url, imagePath, imageData, config.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to upload image: %v\n", err)
		os.Exit(1)
	}

	// 解析响应数据并输出图片链接
	responseJSON := ImageUploadResponse{}
	err = json.Unmarshal(resBody, &responseJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid response JSON: %v\n", string(resBody))
		os.Exit(1)
	}
	if responseJSON.Success {
		fmt.Printf("%v\n", responseJSON.URL)
	} else {
		fmt.Fprintf(os.Stderr, "Failed to upload image: %v\n", responseJSON.Message)
	}
}

func uploadImage(url string, filePath string, imageData []byte, toekn string) ([]byte, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("upload_file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	part.Write(imageData)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	token := toekn
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
