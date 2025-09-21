package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff 命令执行失败，退出码 %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return "", fmt.Errorf("执行 git diff 命令失败: %w", err)
	}

	diff := strings.TrimSpace(string(output))
	if diff == "" {
		return "", fmt.Errorf("没有找到暂存的更改")
	}

	return diff, nil
}

func ValidateGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("不是一个 git 仓库")
	}
	return nil
}
