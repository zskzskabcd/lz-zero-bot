package config

// 配置信息

type BotConfigItem struct {
	// 机器人模式
	Role       string         `json:"role"`
	RoleConfig RoleConfigItem `json:"role_config"`
}

type RoleConfigItem struct {
	// 最大上下文prompt
	MaxPrompt int `json:"max_prompt"`
	// 最大对话条数
	MaxDialogCount int `json:"max_dialog_count"`
	// 最大回答长度
	MaxAnswer int `json:"max_answer"`
	// 基础prompt
	Prompt string `json:"prompt"`
}

type ChatGPTConfig struct {
	// APIKey
	APIKey string `json:"api_key"`
}

type Config struct {
	// 机器人配置 每个群组或者用户一个配置
	BotConfig map[int64]BotConfigItem `json:"bot_config"`
	// 机器人qq号
	BotQQ int64 `json:"bot_qq"`
	// CallBack 配置
	CallBackPrefix string `json:"callback_prefix"`
	// ChatGPT 配置
	ChatGPT ChatGPTConfig `json:"chat_gpt"`
}
