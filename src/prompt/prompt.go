package prompt

import (
	"fmt"
	"strings"
)

const DefaultPrompt = `你是一个 Git 提交信息生成器。请根据以下暂存的更改，请使用中文生成符合 Conventional Commits 规范的提交信息。

要求:
1. 格式: <type>(<scope>): <description>
2. 类型: feat, fix, docs, style, refactor, test, chore
3. 保持第一行少于 50 个字符
4. 如果需要，添加一个空行和详细描述
5. 如果适用，引用相关问题 (例如: "Fixes: #123")
6. 使用现在时态 ("add" 而不是 "added")

暂存的更改:
%s

只生成提交信息，不要包含任何额外的解释或 Markdown 格式:`

func BuildPrompt(diff string) string {
	return fmt.Sprintf(DefaultPrompt, diff)
}

func SanitizeCommitMessage(message string) string {
	// 清理 LLM 返回的消息，移除不必要的格式
	lines := strings.Split(strings.TrimSpace(message), "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 移除 markdown 格式标记
		line = strings.TrimPrefix(line, "```")
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")

		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}
