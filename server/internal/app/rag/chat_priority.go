package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	appsysconfig "github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	appconfig "github.com/shawns-yao/shawn-blog/server/internal/config"
	domainconfig "github.com/shawns-yao/shawn-blog/server/internal/domain/config"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

const chatPriorityKey = "rag.chatPriority"

var defaultChatPriority = []string{"gpt", "grok", "gemini", "opencode_go"}

func validateChatPriority(priority []string) error {
	if len(priority) != len(defaultChatPriority) {
		return fmt.Errorf("请为 GPT、Grok、Gemini、OpenCode Go 四个通道设置完整顺序。")
	}
	seen := make(map[string]bool, len(priority))
	for _, name := range priority {
		if !slices.Contains(defaultChatPriority, name) || seen[name] {
			return fmt.Errorf("通道不能重复或包含未知通道；default 兜底固定在最后。")
		}
		seen[name] = true
	}
	return nil
}

func decodeChatPriority(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return slices.Clone(defaultChatPriority), nil
	}
	var priority []string
	if json.Unmarshal([]byte(value), &priority) != nil || validateChatPriority(priority) != nil {
		return nil, fmt.Errorf("语言模型优先级配置无效，请重新保存。")
	}
	return priority, nil
}

func (s *Service) orderedChatChannels(priority []string) []appconfig.RAGChatConfig {
	byName := make(map[string]appconfig.RAGChatConfig, len(defaultChatPriority))
	for _, channel := range s.providers.ChatChannels() {
		byName[channel.Name] = channel
	}
	channels := make([]appconfig.RAGChatConfig, 0, len(priority)+1)
	for _, name := range priority {
		channels = append(channels, byName[name])
	}
	// The official fallback is outside the editable priority list.
	return append(channels, s.providers.Fallback)
}

func (s *Service) UpdateChatPriority(ctx context.Context, priority []string) (domain.AdminSettings, error) {
	if err := validateChatPriority(priority); err != nil {
		return domain.AdminSettings{}, err
	}
	writer, ok := s.config.(interface {
		UpdateConfigs(context.Context, []appsysconfig.UpdateItem) ([]domainconfig.SysConfig, error)
	})
	if !ok {
		return domain.AdminSettings{}, fmt.Errorf("配置服务不支持写入。")
	}
	value, _ := json.Marshal(priority)
	raw := json.RawMessage(value)
	valueType, group, label := "json", "rag", "语言模型优先级"
	description := "四个可调整通道的调用顺序，DeepSeek 官方 default 固定兜底"
	if _, err := writer.UpdateConfigs(ctx, []appsysconfig.UpdateItem{{Key: chatPriorityKey, Value: &raw,
		ValueType: &valueType, GroupPath: &group, Label: &label, Description: &description}}); err != nil {
		return domain.AdminSettings{}, fmt.Errorf("语言模型优先级保存失败。")
	}
	return s.AdminSettings(ctx)
}
