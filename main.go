package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

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
	config, err := loadConfig()
	if err != nil {
		return
	}
	// 准备上传数据
	imagePath := os.Args[len(os.Args)-1]
	// 判断传入的URL是否是网络图片
	if imagePath[:4] == "http" {
		// 获取当前程序所在目录路径
		executablePath, err := os.Executable()
		if err != nil {
			fmt.Println("Error:", err)
		}
		// 判断当前目录下是否存在temp文件夹，不存在则创建
		executableDir := filepath.Dir(executablePath)
		tempDir := filepath.Join(executableDir, "temp")
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			err := os.Mkdir(tempDir, os.ModePerm)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
		// 获取当前时间戳并转为字符串
		t := time.Now().Unix()
		timestamp := strconv.FormatInt(t, 10)
		// 如果是网络图片先将图片下载到本地
		resp, err := http.Get(imagePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download image: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		imageData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read image: %v\n", err)
			os.Exit(1)
		}
		// 从URL中解析出文件名
		u, err := url.Parse(imagePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse URL: %v\n", err)
			os.Exit(1)
		}
		filename := path.Base(u.Path)
		// 获取文件后缀
		suffix := filepath.Ext(filename)
		// 拼接temp路径

		// 拼接保存路径图片名称加上时间戳，再将图片保存到本地
		imagePath = filepath.Join(tempDir, timestamp+"_temp"+suffix)
		err = ioutil.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save image: %v\n", err)
			os.Exit(1)
		}
	}
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
