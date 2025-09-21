package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gac-core/src/config"
	"gac-core/src/git"
	"gac-core/src/llm"
	"gac-core/src/prompt"
)

func main() {
	log.SetOutput(os.Stderr)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		if err := installHook(); err != nil {
			log.Printf("错误: %v", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("gac-core v1.0.0")
	default:
		// 默认行为：处理 Git Hook 调用
		if err := run(os.Args[1]); err != nil {
			log.Printf("错误: %v", err)
			os.Exit(1)
		}
	}
}

func run(commitMsgFile string) error {
	if err := git.ValidateGitRepo(); err != nil {
		return fmt.Errorf("git 仓库验证失败: %w", err)
	}

	// 检查是否需要跳过
	const skipKeyword = "[skip]"
	content, err := os.ReadFile(commitMsgFile)
	if err == nil {
		contentStr := strings.ToLower(string(content))
		if strings.HasPrefix(contentStr, skipKeyword) {
			// 移除关键字并写回文件
			newMessage := strings.TrimSpace(string(content)[len(skipKeyword):])
			if err := os.WriteFile(commitMsgFile, []byte(newMessage), 0644); err != nil {
				return fmt.Errorf("写入提交信息文件失败: %w", err)
			}
			return nil
		}

		// 如果消息已经是 conventional commit 格式，则直接通过
		if isConventionalCommit(contentStr) {
			return nil
		}
	}

	diff, err := git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("获取暂存区的变更失败: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.LLM.APIKey == "" {
		return fmt.Errorf("API 密钥未配置。请设置 GAC_API_KEY 环境变量或将其添加到 .gac_config.json 文件中")
	}

	client := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.Endpoint, cfg.LLM.Model)

	promptText := prompt.BuildPrompt(diff)

	commitMessage, err := client.GenerateCommitMessage(promptText)
	if err != nil {
		return fmt.Errorf("生成提交信息失败: %w", err)
	}

	cleanMessage := prompt.SanitizeCommitMessage(commitMessage)

	fmt.Println("---------- AI 生成的 Commit Message ----------")
	fmt.Println(cleanMessage)
	fmt.Println("------------------------------------------")

	if err := os.WriteFile(commitMsgFile, []byte(cleanMessage), 0644); err != nil {
		return fmt.Errorf("写入提交信息文件失败: %w", err)
	}

	// 中断提交，让用户审核
	return fmt.Errorf("已生成 commit message，请检查后重新提交")
}

func isConventionalCommit(msg string) bool {
	prefixes := []string{
		"feat:", "fix:", "chore:", "docs:", "style:",
		"refactor:", "perf:", "test:", "build:", "ci:",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(msg, p) {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Printf(`GAC (Git AI Commit) - AI 驱动的提交信息生成器

用法:
  %s <commit-msg-file>    处理提交信息 (由 Git 钩子调用)
  %s install             在当前仓库中安装 Git 钩子
  %s version             显示版本信息

环境变量:
  GAC_API_KEY             LLM API 密钥 (必需)

配置:
  ~/.gac/config.json      全局配置
  ./.gac_config.json      项目级配置 (覆盖全局配置)

`, os.Args[0], os.Args[0], os.Args[0])
}

func installHook() error {
	if err := git.ValidateGitRepo(); err != nil {
		return fmt.Errorf("不在 Git 仓库中: %w", err)
	}

	hookContent := `#!/bin/sh
# GAC (Git AI Commit) Hook -- DEBUG MODE

# 1. Setup Log File
LOG_FILE="$HOME/.gac_hook.log"

# 2. Clear old log and add a separator for the new run
echo "\n--- GAC HOOK START ---" >> "$LOG_FILE"

# 3. Log basic info
date >> "$LOG_FILE"
echo "Args: COMMIT_MSG_FILE='$1' COMMIT_SOURCE='$2' SHA1='$3'" >> "$LOG_FILE"

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2

# 4. Run GAC only in specific cases
if [ -z "$COMMIT_SOURCE" ] || [ "$COMMIT_SOURCE" = "message" ] || [ "$COMMIT_SOURCE" = "template" ]; then
    echo "Condition met: Running GAC..." >> "$LOG_FILE"

    # 5. Find the gac-core executable
    GAC_EXEC=""
    if command -v gac-core.exe >/dev/null 2>&1; then GAC_EXEC="gac-core.exe";
    elif command -v gac-core >/dev/null 2>&1; then GAC_EXEC="gac-core";
    elif [ -x "$HOME/.gac/bin/gac-core.exe" ]; then GAC_EXEC="$HOME/.gac/bin/gac-core.exe";
    elif [ -x "$HOME/.local/bin/gac-core.exe" ]; then GAC_EXEC="$HOME/.local/bin/gac-core.exe";
    elif [ -x "$HOME/.local/bin/gac-core" ]; then GAC_EXEC="$HOME/.local/bin/gac-core";
    elif [ -x "./gac-core.exe" ]; then GAC_EXEC="./gac-core.exe";
    elif [ -x "./src/gac-core.exe" ]; then GAC_EXEC="./src/gac-core.exe";
    elif [ -x "./src/gac-core" ]; then GAC_EXEC="./src/gac-core";
    fi

    # 6. Execute GAC if found and handle exit code
    if [ -n "$GAC_EXEC" ]; then
        echo "Found GAC_EXEC: '$GAC_EXEC'" >> "$LOG_FILE"
        
        # Read commit message content before running GAC
        echo "Content of '$COMMIT_MSG_FILE' before GAC run:" >> "$LOG_FILE"
        cat "$COMMIT_MSG_FILE" >> "$LOG_FILE"

        # Execute, redirecting stdout and stderr to the log file
        "$GAC_EXEC" "$COMMIT_MSG_FILE" >> "$LOG_FILE" 2>&1
        STATUS=$?
        
        echo "gac-core exited with status: $STATUS" >> "$LOG_FILE"
        
        # Read commit message content after running GAC
        echo "Content of '$COMMIT_MSG_FILE' after GAC run:" >> "$LOG_FILE"
        cat "$COMMIT_MSG_FILE" >> "$LOG_FILE"

        echo "Exiting hook with status $STATUS." >> "$LOG_FILE"
        exit $STATUS
    else
        echo "GAC_EXEC not found." >> "$LOG_FILE"
    fi
else
    echo "Condition not met. Skipping GAC." >> "$LOG_FILE"
fi

# 7. If GAC was not run, exit 0 to allow the commit to proceed.
echo "Exiting hook with status 0." >> "$LOG_FILE"
exit 0`

	hookPath := ".git/hooks/prepare-commit-msg"

	if err := os.MkdirAll(".git/hooks", 0755); err != nil {
		return fmt.Errorf("创建钩子目录失败: %w", err)
	}

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("写入钩子文件失败: %w", err)
	}

	fmt.Println("✓ Git 钩子安装成功！(调试模式)")
	fmt.Println("✓ GAC 现在将自动生成提交信息")

	return nil
}
