package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grtsinry43/grtblog/cli/v2/internal/client"
)

// taxonomyIndex 缓存专栏/标签，用于名称 ↔ ID 解析。
type taxonomyIndex struct {
	Columns []client.Column
	Tags    []client.Tag
}

func loadTaxonomy(c *client.Client) (*taxonomyIndex, error) {
	idx := &taxonomyIndex{}
	if err := c.Get("/columns", nil, &idx.Columns); err != nil {
		return nil, err
	}
	if err := c.Get("/tags", nil, &idx.Tags); err != nil {
		return nil, err
	}
	return idx, nil
}

// resolveID 将「名称或数字 ID」解析为数字 ID。
func resolveID(nameOrID string, names map[string]int64, kind string) (int64, error) {
	s := strings.TrimSpace(nameOrID)
	if s == "" {
		return 0, fmt.Errorf("%s不能为空", kind)
	}
	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		return id, nil
	}
	if id, ok := names[strings.ToLower(s)]; ok {
		return id, nil
	}
	available := make([]string, 0, len(names))
	for n := range names {
		available = append(available, n)
	}
	sort.Strings(available)
	return 0, fmt.Errorf("未找到%s %q（可用: %s）", kind, s, strings.Join(available, ", "))
}

func (t *taxonomyIndex) columnNames() map[string]int64 {
	m := make(map[string]int64, len(t.Columns))
	for _, c := range t.Columns {
		m[strings.ToLower(c.Name)] = c.ID
	}
	return m
}

func (t *taxonomyIndex) tagNames() map[string]int64 {
	m := make(map[string]int64, len(t.Tags))
	for _, tag := range t.Tags {
		m[strings.ToLower(tag.Name)] = tag.ID
	}
	return m
}

// parseIDs 将命令行参数解析为 int64 ID 列表。
func parseIDs(args []string) ([]int64, error) {
	ids := make([]int64, 0, len(args))
	for _, a := range args {
		id, err := strconv.ParseInt(strings.TrimSpace(a), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("无效的 ID %q", a)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// resolveIDs 批量解析名称/ID 混合列表。
func resolveIDs(list []string, names map[string]int64, kind string) ([]int64, error) {
	if len(list) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(list))
	for _, item := range list {
		id, err := resolveID(item, names, kind)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// contentStatus 渲染手记发布状态（纯文本，供表格与详情共用）。
func contentStatus(published, top bool) string {
	switch {
	case published && top:
		return "已发布·置顶"
	case published:
		return "已发布"
	case top:
		return "草稿·置顶"
	default:
		return "草稿"
	}
}

// coverMark 渲染封面有无标记。
func coverMark(cover *string) string {
	if cover == nil || strings.TrimSpace(*cover) == "" {
		return "-"
	}
	return "有"
}

// strPtr 将非空字符串转为指针，空字符串返回 nil。
func strPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// readContent 从参数拼接内容；无参数时从标准输入读取。
func readContent(args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("读取标准输入失败: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// batchSpec 描述一个批量操作子命令。
type batchSpec struct {
	use     string
	short   string
	confirm string // 非空时执行前询问确认，%d 替换为 ID 数量
	run     func(cli *client.Client, ids []int64) error
}

// newBatchCmd 构造批量操作命令：解析 ID 列表、可选确认、执行。
func newBatchCmd(app *App, spec batchSpec) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := parseIDs(args)
			if err != nil {
				return err
			}
			cli, err := app.Client()
			if err != nil {
				return err
			}
			if spec.confirm != "" && !yes {
				if !app.Out.Confirm(spec.confirm, len(ids)) {
					app.Out.Warnf("已取消")
					return nil
				}
			}
			return spec.run(cli, ids)
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "跳过确认")
	return cmd
}
