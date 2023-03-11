package ChatAI

import (
	"bytes"
	"encoding/json"
	chatAI "github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/config"
	log2 "github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const BASEURL = "https://api.openai.com/v1/"

// GptResponseBody 请求体
type GptResponseBody struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int                      `json:"created"`
	Model   string                   `json:"model"`
	Choices []map[string]interface{} `json:"choices"`
	Usage   map[string]interface{}   `json:"usage"`
}

type ChoiceItem struct {
}

// GptRequestBody 请求体
type GptRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

// ChatGPTRequestMessage chatGPT请求体
type ChatGPTRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTRequestBody chatGPT请求体
type ChatGPTRequestBody struct {
	Model    string                  `json:"model"`
	Messages []ChatGPTRequestMessage `json:"messages"`
}

type ChatGPTResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// TextCompletions gtp文本模型回复
func TextCompletions(reqConf GptRequestBody) (string, error) {
	requestBody := reqConf
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := chatAI.GetChatGPTConfig().APIKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	gptResponseBody := &GptResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}
	var reply string
	if len(gptResponseBody.Choices) > 0 {
		for _, v := range gptResponseBody.Choices {
			reply = v["text"].(string)
			break
		}
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply, nil
}

// ChatCompletions gpt 生成对话
func ChatCompletions(reqConf ChatGPTRequestBody) (string, error) {
	requestBody := reqConf
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := chatAI.GetChatGPTConfig().APIKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{
		Timeout: 100 * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log2.Print(err)
		return "", err
	}

	gptResponseBody := &ChatGPTResponse{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}
	var reply string
	if len(gptResponseBody.Choices) > 0 {
		for _, v := range gptResponseBody.Choices {
			reply = v.Message.Content
			break
		}
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply, nil
}

// CodeCompletion gpt 代码编写
func CodeCompletion(msg string) (string, error) {
	req := GptRequestBody{
		Model:            "code-davinci-002",
		Prompt:           msg,
		MaxTokens:        2048,
		Temperature:      0.3,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	res, err := TextCompletions(req)
	if err != nil {
		return "", err
	}
	// 逐行遍历 去除首位的+号
	var code string
	for _, v := range strings.Split(res, "\n") {
		code += strings.TrimLeft(v, "+") + "\n"
	}
	return code, nil
}

// TextCompletion gpt 生成文本
func TextCompletion(msg string, maxToken int) (string, error) {
	req := GptRequestBody{
		Model:            "text-davinci-003",
		Prompt:           msg,
		MaxTokens:        maxToken,
		Temperature:      0.5,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	return TextCompletions(req)
}

// ChatCompletion gpt 聊天模型回复
func ChatCompletion(msg []ChatGPTRequestMessage) (string, error) {
	req := ChatGPTRequestBody{
		Model:    "gpt-3.5-turbo",
		Messages: msg,
	}
	return ChatCompletions(req)
}
