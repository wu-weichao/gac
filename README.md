# GAC (Git AI Commit) - AI驱动的Git提交信息生成器

GAC 是一个基于大语言模型的 Git 提交信息生成工具，能够自动分析你的代码变更并生成符合 Conventional Commits 规范的提交信息。

## ✨ 特性

- 🤖 **AI 驱动**: 使用大语言模型分析代码变更生成提交信息
- 📋 **规范格式**: 自动生成符合 Conventional Commits 规范的提交信息  
- 🔧 **无缝集成**: 通过 Git hooks 集成，对现有工作流无侵入
- 🌍 **全局安装**: 一次安装，多项目使用
- ⚙️ **灵活配置**: 支持多种 LLM API，可自定义配置
- 🔒 **安全私密**: 本地处理，仅通过 HTTPS 与 LLM 服务通信

## 🚀 快速开始

### 安装

**Linux/macOS:**【待测试】
```bash
curl -fsSL https://raw.githubusercontent.com/wu-weichao/gac/main/install.sh | bash
```

**Windows:**
```powershell
# 下载并运行安装脚本
curl -fsSL https://raw.githubusercontent.com/wu-weichao/gac/main/install.bat -o install.bat && install.bat
```


**从源码安装 (需要 Go 1.18+):**
```bash
git clone https://github.com/wu-weichao/gac.git
cd gac
chmod +x install.sh
./install.sh
```

### 配置

1. **设置 API Key:**
```bash
export GAC_API_KEY=your-openai-api-key
```

2. **或编辑配置文件:**
```bash
# 全局配置
~/.gac/config.json

# 项目级配置 (会覆盖全局配置)
./.gac_config.json
```

配置文件格式：
```json
{
  "llm": {
    "provider": "openai",
    "api_key": "your-api-key",
    "endpoint": "https://api.openai.com/v1/chat/completions",
    "model": "gpt-3.5-turbo"
  }
}
```

### 使用

1. **在任意 Git 仓库中安装 Hook:**
```bash
gac-core install
```

2. **正常提交代码:**
```bash
git add .
git commit  # 提交信息会自动生成
```

## 📖 使用说明

### 命令

- `gac-core <commit-msg-file>` - 处理提交信息 (由 Git hook 调用)
- `gac-core install` - 在当前仓库安装 Git hook
- `gac-core version` - 显示版本信息

### 工作流程

1. 你在 IDE 中点击提交或运行 `git commit`
2. Git 自动触发 `prepare-commit-msg` hook
3. GAC 获取暂存的代码变更 (`git diff --staged`)
4. 将变更发送给 LLM 生成提交信息
5. 生成的提交信息自动填充到提交框中
6. 你可以确认或修改后完成提交

### 配置优先级

1. 项目级配置 (`./.gac_config.json`)
2. 全局配置 (`~/.gac/config.json`)
3. 环境变量 (`GAC_API_KEY`)
4. 默认配置

### 支持的 LLM 提供商

- OpenAI (GPT-3.5/GPT-4)
- 其他兼容 OpenAI API 格式的服务商

## 🔧 高级配置

### 自定义 Prompt

你可以通过修改源码中的 `prompt.DefaultPrompt` 来自定义生成逻辑，或者扩展配置文件支持自定义 prompt。

### 多项目不同配置

每个项目可以有自己的 `.gac_config.json` 配置文件，允许不同项目使用不同的模型或 API Key。

## 📄 许可证

MIT License

## 功能
- [x] 基于 Git 钩子 (Git Hooks) 的客户端工具生成 Commit
- [ ] IDE 插件扩展

## ❓ 常见问题

**Q: 如何卸载？**
A: 删除 `~/.gac` 目录和 `~/.local/bin/gac-core`，并从各项目中删除 `.git/hooks/prepare-commit-msg`。

**Q: 为什么提交信息没有自动生成？**
A: 
1. 检查是否已设置 API Key
2. 检查是否有暂存的文件变更
3. 查看错误日志 (stderr)

**Q: 支持私有部署的 LLM 吗？**
A: 支持，只需修改配置文件中的 `endpoint` 字段指向你的 API 地址。

**Q: 如何禁用某次提交的自动生成？**
A: 有多种方法可以禁用单次提交的自动生成功能：

1. **使用关键词禁用** (推荐，尤其适合 IDE 用户):
   在 commit 信息的第一行输入 `[skip]`，GAC 将会跳过生成，并自动移除该关键词。
   例如: `[skip] fix: 这是一个手写的提交`

2. **重命名/删除钩子文件**：暂时重命名或删除 `.git/hooks/prepare-commit-msg` 文件
   ```bash
   mv .git/hooks/prepare-commit-msg .git/hooks/prepare-commit-msg.bak  # 临时重命名
   git commit                                                           # 执行提交
   mv .git/hooks/prepare-commit-msg.bak .git/hooks/prepare-commit-msg  # 恢复钩子文件
   ```