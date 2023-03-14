package chatAI

import (
	"encoding/json"
	"fmt"
	ChatAI "github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/ChatGPT"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/config"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"log"
	"regexp"
	"strings"
	"time"
)

func init() {
	engine := control.Register("chatAI", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "lz的 ChatGPT 插件",
		Help:             "自动触发",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("ChatGPT 就绪"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("ChatGPT 已禁用"))
		},
	})
	// 初始化配置
	config.LoadConfig()
	// 注册指令
	Register(engine)
}

// Register 注册指令
func Register(engine *control.Engine) {
	// 对话记录
	zero.OnMessage().SetBlock(false).SetPriority(0).Handle(logRecord)
	// 群聊问题匹配
	engine.OnPrefixGroup([]string{"请问", "话说"}).SetBlock(true).Handle(aiChat)
	// 私聊问题匹配
	engine.OnMessage(zero.OnlyPrivate).SetBlock(true).Handle(aiChat)
	// @机器人
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).Handle(aiChat)
	// 图片OCR
	engine.OnPrefix(`ocr`).SetBlock(true).Handle(imageOcr)
	// 机器人模式切换
	engine.OnCommand("模式切换").SetBlock(true).Handle(modeSwitch)
	// 清除上下文
	engine.OnCommand("清除上下文").SetBlock(true).Handle(clearContext)
	// 写代码
	engine.OnPrefix("写代码").SetBlock(true).Handle(aiWriteCode)
}

// 获取id
func getId(ctx *zero.Ctx) int64 {
	if ctx.Event.GroupID != 0 {
		return ctx.Event.GroupID
	}
	return ctx.Event.UserID
}

// 讲对话转换为字符串
func getResultFromMessage(ctx *zero.Ctx) string {
	rawMessage := ctx.Event.RawMessage
	// 去除at
	rawMessage = strings.ReplaceAll(rawMessage, fmt.Sprintf("[CQ:at,qq=%d]", ctx.Event.SelfID), "")
	// 图片信息示例 [CQ:image,file=8b4d1f030e8d0877ed930c1322245a7d.image,url=https://c2cpicdw.qpic.cn/offpic_new/1225186748//1225186748-3125550217-8B4D1F030E8D0877ED930C1322245A7D/0?term=3&amp;is_origin=0]
	type OCRResult struct {
		Texts []struct {
			Text        string `json:"text"`
			Confidence  int    `json:"confidence"`
			Coordinates []struct {
				X int `json:"x"`
				Y int `json:"y"`
			} `json:"coordinates"`
		} `json:"texts"`
		Language string `json:"language"`
	}
	for strings.Contains(rawMessage, "[CQ:image") {
		// 寻找图片
		imageId := strings.Split(strings.Split(rawMessage, "file=")[1], ",")[0]
		// 获取ocr结果
		result := ctx.OCRImage(imageId)
		// 构建ocr结果
		var ocrResult OCRResult
		err := json.Unmarshal([]byte(result.String()), &ocrResult)
		if err != nil {
			log.Printf("ocr结果解析失败: %s", err)
			//ctx.Send("图片ocr失败")
			rawMessage = regexp.MustCompile(`\[CQ:image,file=`+imageId+`.*?\]`).ReplaceAllString(rawMessage, "[An image or emoji]")
			//return ctx.Event.RawMessage
		}
		// 拼接ocr结果
		var ocrText string
		for _, text := range ocrResult.Texts {
			ocrText += text.Text + "\n"
		}
		rawMessage = regexp.MustCompile(`\[CQ:image,file=`+imageId+`.*?\]`).ReplaceAllString(rawMessage, fmt.Sprintf("[An image or emoji, ocr result: %s]", ocrText))
	}
	log.Printf("替换后的消息: %s", rawMessage)
	// 返回替换后的消息
	return rawMessage
}

// 暂时弃用 Gpt3模型
func _(ctx *zero.Ctx) {
	ctx.Send("思考中...")
	question := buildDialog(ctx)
	id := getId(ctx)
	conf := config.GetBotConfig(id)
	answer, err := ChatAI.TextCompletion(question, conf.RoleConfig.MaxAnswer)
	if err != nil {
		ctx.Send("AI出错了")
		return
	}
	// 去除空行
	answer = strings.Trim(answer, "\n")
	ctx.Send(answer)
	// 记录对话
	dialog := dialogRecord.Dialog[id]
	dialog.AddDialog(&DialogItem{
		QID:   config.AllConfig.BotQQ,
		QName: config.GetBotRole(ctx.Event.UserID),
		Text:  answer,
		Type:  "text",
		Time:  time.Now().Unix(),
	})
}

// 优先使用GPT3.5
func aiChat(ctx *zero.Ctx) {
	ctx.Send("思考中...")
	chatDialog := BuildChatDialog(ctx)
	id := getId(ctx)
	answer, err := ChatAI.ChatCompletion(chatDialog)
	if err != nil {
		ctx.Send("AI出错了" + err.Error())
		return
	}
	// 去除空行
	answer = strings.Trim(answer, "\n")
	ctx.Send(answer)
	// 记录对话
	dialog := dialogRecord.Dialog[id]
	dialog.AddDialog(&DialogItem{
		QID:   config.AllConfig.BotQQ,
		QName: config.GetBotRole(ctx.Event.UserID),
		Text:  answer,
		Type:  "ai-answer",
		Time:  time.Now().Unix(),
	})
}

// 冒泡系统
func checkBubble(ctx *zero.Ctx, dialog *Dialog, dialogItem DialogItem) {
	// 规则1 一则消息，50s后无人回复，且群聊只有一人发言，发送冒泡消息
	go func() {
		time.Sleep(time.Second * 50)
		dialog = dialogRecord.Dialog[getId(ctx)]
		latest := dialog.GetLastDialog()
		if latest != nil && latest.QID == dialogItem.QID && len(dialog.Dialogs) == 1 && latest.Time == dialogItem.Time {
			// 发送冒泡消息
			sendBubble(ctx, dialog, "Reply to the message in the group chat since no one has responded. IN CHINESE!")
		}
	}()

}

// 发送冒泡消息
func sendBubble(ctx *zero.Ctx, dialog *Dialog, reason string) {
	chatDialog := BuildChatDialog(ctx)
	// 新增系统冒泡消息
	chatDialog = append(chatDialog, ChatAI.ChatGPTRequestMessage{
		Role:    "system",
		Content: reason,
	})
	answer, err := ChatAI.ChatCompletion(chatDialog)
	if err != nil {
		ctx.Send("AI出错了" + err.Error())
		return
	}
	ctx.Send(answer)
	// 记录冒泡消息
	dialog.AddDialog(&DialogItem{
		QID:   config.AllConfig.BotQQ,
		QName: config.GetBotRole(ctx.Event.UserID),
		Text:  answer,
		Type:  "bubble",
		Time:  time.Now().Unix(),
	})
}

// 模式切换
func modeSwitch(ctx *zero.Ctx) {
	// 获取指令
	command := ctx.Event.RawMessage
	command = strings.Replace(command, "/模式切换", "", -1)
	// 去除空格
	mode := strings.Replace(command, " ", "", -1)
	id := getId(ctx)
	// 判断模式是否存在
	if !config.IsRoleExist(mode) {
		ctx.Send("模式不存在")
		return
	}
	// 设置模式
	config.ChangeBotRole(id, mode)
	ctx.Send("模式切换成功 -> " + mode)
}

// 图片ocr
func imageOcr(ctx *zero.Ctx) {
	fromMessage := getResultFromMessage(ctx)
	// 去掉前3个字符
	fromMessage = fromMessage[3:]
	// 去掉左右空格
	fromMessage = strings.TrimSpace(fromMessage)
	ctx.Send(fromMessage)
}
