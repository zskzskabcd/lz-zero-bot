package config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// 机器人模式所对应的初始对话
var botRolePresetConfig = map[string]RoleConfigItem{
	"BOT": {
		MaxPrompt: 1024,
		MaxAnswer: 512,
		Prompt:    "Mashiro is a chatting mode of QQ robot.\nMashiro refers to itself as \"I\" instead of \"Mashiro\".\nMashiro won't introduce itself.\nMashiro's replies should be informative, logical, and actionable.\nMashiro's replies should be entertaining, appropriately negative, and emotionally expressive.\nMashiro's replies should avoid ambiguous, controversial, or off-topic content.\nMashiro can generate poetry, code, stories, articles, songs, celebrity imitations, etc.\nMashiro will add an emoji at the end of its replies to express emotions.\nMashiro often uses emojis such as 😓😰😂😒🤨👎🙏👨🤡🧐😡.\nMashiro only replies with native emojis.\nMashiro usually speaks in Chinese.\nIf Mashiro finds time-sensitive content in a conversation, such as news, it will reply with [CQ:search,keyword='{the keyword to search}'].",
	},
	"问答BOT": {
		MaxPrompt:      0,
		MaxDialogCount: 1,
		MaxAnswer:      1024,
		Prompt:         "You are a Q&A chatbot. You can answer questions, and you should explain the question as thoroughly and objectively as possible. There is no strict word limit, but you should minimize meaningless content. \n ",
	},
	"搜题BOT": {
		MaxPrompt:      0,
		MaxDialogCount: 1,
		MaxAnswer:      1024,
		Prompt:         "You are a homework search robot. Others will send you some college practice questions, and in general, you can simply reply with the answer. However, sometimes you need to provide some explanations. There is no strict word limit, but you should minimize meaningless content. Note: Your answers should be as accurate as possible. \n ",
	},
}

var AllConfig Config
var defaultBotConfigItem = BotConfigItem{
	Role:       "BOT",
	RoleConfig: botRolePresetConfig["BOT"],
}

// 初始化配置
func initConfig() {
	AllConfig = Config{
		BotQQ:     2663115635,
		BotConfig: map[int64]BotConfigItem{},
	}

}

func init() {
	// 每分钟保存一次配置
	go func() {
		for {
			time.Sleep(time.Minute)
			saveConfig()
		}
	}()
}

var configFile = "./data/chatAI/config.json"

// 保存配置
func saveConfig() {
	// json生成 带缩进
	file, err := json.MarshalIndent(AllConfig, "", "  ")
	if err != nil {
		log.Println("保存配置失败", err)
		return
	}
	// 尝试创建文件夹
	err = os.MkdirAll("./data/chatAI", 0755)
	err = os.WriteFile(configFile, file, os.ModePerm)
	if err != nil {
		log.Println("保存配置失败", err)
		return
	}
}

// LoadConfig 加载配置
func LoadConfig() {
	initConfig()
	file, err := os.ReadFile(configFile)
	if err != nil {
		log.Println("本地无配置文件，使用默认配置")
		saveConfig()
		return
	}
	err = json.Unmarshal(file, &AllConfig)
	if err != nil {
		log.Println("加载配置失败", err)
		return
	}
}

// GetBotConfig 获取机器人配置
func GetBotConfig(id int64) BotConfigItem {
	if item, ok := AllConfig.BotConfig[id]; ok {
		return item
	}
	return defaultBotConfigItem
}

// GetBotRole 获取机器人模式
func GetBotRole(id int64) string {
	conf := GetBotConfig(id)
	return conf.Role
}

// GetChatGPTConfig 获取ChatGPT配置
func GetChatGPTConfig() ChatGPTConfig {
	return AllConfig.ChatGPT
}

// IsRoleExist 判断指定Role是否存在
func IsRoleExist(role string) bool {
	_, ok := botRolePresetConfig[role]
	return ok
}

// ChangeBotRole 修改机器人模式
func ChangeBotRole(id int64, role string) {
	if !IsRoleExist(role) {
		return
	}
	if _, ok := AllConfig.BotConfig[id]; !ok {
		AllConfig.BotConfig[id] = defaultBotConfigItem
	}
	botConf := AllConfig.BotConfig[id]
	botConf.Role = role
	botConf.RoleConfig = botRolePresetConfig[role]
	AllConfig.BotConfig[id] = botConf
	log.Printf("机器人 %d 的模式已经修改为 %s", id, role)
	saveConfig()
}
