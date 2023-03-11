package chatAI

import (
	"fmt"
	ChatAI "github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/ChatGPT"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/chat_ai/utils/config"
	zero "github.com/wdvxdr1123/ZeroBot"
	"strings"
	"time"
)

type DialogItem struct {
	QID       int64
	QName     string
	MessageID int64
	Text      string
	Type      string
	Time      int64
}

type Dialog struct {
	// 消息记录
	Dialogs []DialogItem `json:"dialog"`
	ID      int64        `json:"id"`
	Type    string       `json:"type"`
}

type DialogRecord struct {
	Dialog map[int64]*Dialog `json:"dialogs"`
}

var dialogRecord DialogRecord

func init() {
	// 初始化
	dialogRecord = DialogRecord{
		Dialog: make(map[int64]*Dialog),
	}

	// 每隔1分钟清理一次对话记录
	go func() {
		for {
			time.Sleep(time.Minute)
			cleanDialog()
		}
	}()
}

// 消息记录

// AddDialog 新增对话记录
func (d *Dialog) AddDialog(dialog *DialogItem) {
	d.Dialogs = append(d.Dialogs, *dialog)
}

// GetLastDialog 获取最后一条对话
func (d *Dialog) GetLastDialog() *DialogItem {
	if len(d.Dialogs) == 0 {
		return nil
	}
	return &d.Dialogs[len(d.Dialogs)-1]
}

// ClearDialog 清除对话记录
func (d *Dialog) ClearDialog() {
	d.Dialogs = []DialogItem{}
}

// GetDialog 获取指定字节长度的对话
func (d *Dialog) GetDialog(maxLen int) string {
	var dialog string
	for i := len(d.Dialogs) - 1; i >= 0; i-- {
		item := d.Dialogs[i]
		// 限制最大字节数 但不限制最后一条
		if len(dialog)+len(item.Text) > maxLen && i != len(d.Dialogs)-1 {
			return dialog
		}
		dialog = fmt.Sprintf("%s: %s\n", item.QName, item.Text) + dialog
	}
	return dialog
}

// GetDialogItem 获取指定字节长度对话记录
func (d *Dialog) GetDialogItem(maxLen int) []DialogItem {
	var dialog []DialogItem
	for i := len(d.Dialogs) - 1; i >= 0; i-- {
		item := d.Dialogs[i]
		// 限制最大字节数 但不限制最后一条
		if len(dialog)+len(item.Text) > maxLen && i != len(d.Dialogs)-1 {
			return dialog
		}
		dialog = append(dialog, item)
	}
	return dialog
}

// GetDialogItemsByTimeSlot 获取指定时间段内的对话
func (d *Dialog) GetDialogItemsByTimeSlot(startTime, endTime int64) []DialogItem {
	var dialog []DialogItem
	for i := len(d.Dialogs) - 1; i >= 0; i-- {
		item := d.Dialogs[i]
		if item.Time < startTime {
			break
		}
		if item.Time > endTime {
			continue
		}
		dialog = append(dialog, item)
	}
	return dialog
}

// GetRecentTimeDialog 获取过去指定时间的对话
func (d *Dialog) GetRecentTimeDialog(pastTime int64) []DialogItem {
	return d.GetDialogItemsByTimeSlot(time.Now().Unix()-pastTime, time.Now().Unix())
}

// GetRecentMinuteDialog 获取过去一分钟的对话
func (d *Dialog) GetRecentMinuteDialog() []DialogItem {
	return d.GetRecentTimeDialog(60)
}

// GetDialogByCount 获取指定条数的对话 限制最大字节数为4k
func (d *Dialog) GetDialogByCount(count int) string {
	var dialog string
	maxLen := 4096
	if len(d.Dialogs) < count {
		count = len(d.Dialogs)
	}
	if count == 0 {
		return ""
	}
	for i := len(d.Dialogs) - 1; i >= 0; i-- {
		item := d.Dialogs[i]
		dialog = fmt.Sprintf("%s: %s\n", item.QName, item.Text) + dialog
		if len(dialog) > maxLen {
			return dialog
		}
		if len(d.Dialogs)-i >= count {
			return dialog
		}
	}
	return dialog
}

func (d *Dialog) GetDialogItemByCount(count int) []DialogItem {
	if len(d.Dialogs) < count {
		count = len(d.Dialogs)
	}
	if count == 0 {
		return []DialogItem{}
	}
	return d.Dialogs[len(d.Dialogs)-count:]
}

func logRecord(ctx *zero.Ctx) {
	// 判断发送内容类型
	msgType := "text"
	// ocr其中的图片
	message := getResultFromMessage(ctx)
	id := getId(ctx)
	//fmt.Printf("记录对话: %d %s", id, message)
	dialog, exist := dialogRecord.Dialog[id]
	if !exist {
		dialog = &Dialog{
			Dialogs: []DialogItem{},
			ID:      id,
			Type:    "group",
		}
		dialogRecord.Dialog[id] = dialog
	}
	dialogItem := DialogItem{
		QID:       ctx.Event.UserID,
		QName:     ctx.Event.Sender.Name(),
		Text:      message,
		Type:      msgType,
		MessageID: ctx.Event.MessageID.(int64),
		Time:      time.Now().Unix(),
	}
	dialog.AddDialog(&dialogItem)

	// 冒泡系统
	go checkBubble(ctx, dialog, dialogItem)
}

// 构建对话
func buildDialog(ctx *zero.Ctx) string {
	var dialog string
	id := getId(ctx)
	conf := config.GetBotConfig(id)
	d, exist := dialogRecord.Dialog[id]
	if !exist {
		d = &Dialog{
			Dialogs: []DialogItem{},
		}
	}
	// 如果对话内容为 “继续” 则获取最近4条对话
	if strings.TrimSpace(ctx.Event.Message.ExtractPlainText()) == "继续" {
		dialog = d.GetDialogByCount(4)
	} else if conf.RoleConfig.MaxPrompt > 0 {
		dialog = d.GetDialog(conf.RoleConfig.MaxPrompt)
	} else if conf.RoleConfig.MaxDialogCount > 0 {
		dialog = d.GetDialogByCount(conf.RoleConfig.MaxDialogCount)
	}
	modePrefix := conf.RoleConfig.Prompt
	dialog = modePrefix + dialog
	// 最后加上回答前缀
	dialog = dialog + fmt.Sprintf("%s: ", conf.Role)
	return dialog
}

// BuildChatDialog 构建对话 - chatGPT
func BuildChatDialog(ctx *zero.Ctx) []ChatAI.ChatGPTRequestMessage {
	var dialog []DialogItem
	var result []ChatAI.ChatGPTRequestMessage
	id := getId(ctx)
	conf := config.GetBotConfig(id)
	d, exist := dialogRecord.Dialog[id]
	if !exist {
		d = &Dialog{
			Dialogs: []DialogItem{},
		}
	}
	// 如果对话内容为 “继续” 则获取最近4条对话
	if strings.TrimSpace(ctx.Event.Message.ExtractPlainText()) == "继续" {
		dialog = d.GetDialogItemByCount(4)
	} else if conf.RoleConfig.MaxPrompt > 0 {
		dialog = d.GetDialogItem(conf.RoleConfig.MaxPrompt)
	} else if conf.RoleConfig.MaxDialogCount > 0 {
		dialog = d.GetDialogItemByCount(conf.RoleConfig.MaxDialogCount)
	}
	// 插入模式前缀
	modePrefix := conf.RoleConfig.Prompt
	result = append(result, ChatAI.ChatGPTRequestMessage{
		Role:    "system",
		Content: modePrefix,
	})
	// 插入系统前缀
	result = append(result, ChatAI.ChatGPTRequestMessage{
		Role:    "system",
		Content: "If replying to someone within a group chat, please add [CQ:at,qq=QID] before the result; To reply to a specific message, add [CQ:reply,id=MID] before the result; Your answer should be in Chinese",
	})
	// 插入时间前缀
	result = append(result, ChatAI.ChatGPTRequestMessage{
		Role:    "system",
		Content: fmt.Sprintf("current time: %s", time.Now().Format("2006-01-02 15:04:05")),
	})
	for i := len(dialog) - 1; i >= 0; i-- {
		item := dialog[i]
		if item.Type == "ai-answer" {
			result = append(result, ChatAI.ChatGPTRequestMessage{
				Role:    "assistant",
				Content: fmt.Sprintf("%s", item.Text),
			})
		} else {
			result = append(result, ChatAI.ChatGPTRequestMessage{
				Role:    "user",
				Content: fmt.Sprintf("[QName:%s QID:%d MID:%d]: %s", item.QName, item.QID, item.MessageID, item.Text),
			})
		}

	}
	return result
}

// 清除对话记录
func clearContext(ctx *zero.Ctx) {
	id := getId(ctx)
	dialog := dialogRecord.Dialog[id]
	dialog.ClearDialog()
	ctx.Send("我已忘记一切")
}

// 执行对话清理
func cleanDialog() {
	// 清理超过1小时的对话
	now := time.Now().Unix()
	for id, dialog := range dialogRecord.Dialog {
		if len(dialog.Dialogs) > 0 && now-dialog.Dialogs[len(dialog.Dialogs)-1].Time > 3600 {
			delete(dialogRecord.Dialog, id)
		}
	}
}
