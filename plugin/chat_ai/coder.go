package chatAI

import (
	"fmt"
	ChatAI "github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/ChatGPT"
	zero "github.com/wdvxdr1123/ZeroBot"
	"strings"
)

// code template

type CodePersist struct {
	Template string `json:"template"`
}

var codePersist = map[string]CodePersist{
	"python": {
		Template: "# -*- coding: utf-8 -*-\"\"\"%s\"\"\"",
	},
	"java": {
		Template: "/*%s*/",
	},
	"javascript": {
		Template: "/*%s*/",
	},
	"c": {
		Template: "/*%s*/",
	},
	"c++": {
		Template: "/*%s*/",
	},
	"go": {
		Template: "/*%s*/",
	},
	"vue": {
		Template: "/*%s*/",
	},
	"html": {
		Template: "<!--%s-->",
	},
	"css": {
		Template: "/*%s*/",
	},
	"default": {
		Template: "/*%s*/",
	},
}

// aiWriteCode 写代码
// 用法：写代码 语言 要求
func aiWriteCode(ctx *zero.Ctx) {
	// 解析参数
	message := getResultFromMessage(ctx)
	// 去除首尾的空格
	message = strings.TrimSpace(message)
	//ctx.Send("message: " + message)
	// 首先 获取第一个空格前的字符串
	args := strings.SplitN(message, " ", 3)
	if len(args) != 3 {
		ctx.Send("参数不足 用法：写代码 语言 要求")
		return
	}
	// 获取语言
	language := args[1]
	// 获取要求
	require := args[2]
	// 获取模板
	template, ok := codePersist[language]
	if !ok {
		template = codePersist["default"]
	}
	msg := fmt.Sprintf(template.Template, require)
	ctx.Send(fmt.Sprintf("模型：%s \n语言：%s \n请求中...", "code-davinci-002", language))
	res, err := ChatAI.CodeCompletion(msg)
	if err != nil {
		ctx.Send("写代码失败 请求时发生错误：" + err.Error())
		return
	}
	// 去除首尾的空格
	res = strings.TrimSpace(res)
	ctx.Send(res)

}
