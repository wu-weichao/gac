# 架构设计文档 - AI Git 提交信息生成器

## 1. 系统概述

本系统是一个基于 Git 钩子 (Git Hooks) 的客户端工具，旨在通过集成大语言模型 (LLM) 的能力，为开发者自动生成 Git 提交信息。其核心设计思想是“无缝集成、本地执行”，通过 `prepare-commit-msg` 钩子在后台完成工作，对开发者的现有工作流无感知、无侵入。

## 2. 架构图

```mermaid
graph TD
    subgraph "开发者本地环境"
        A[开发者在 IDE 中点击 Commit] --> B{Git 触发 a.yaml};
        B --> C[1. prepare-commit-msg 钩子 (Shell)];
        C --> D[2. 核心逻辑程序 (Go Binary)];
        D -- "os/exec: git diff --staged" --> E[获取代码变更];
        D -- "构建 Prompt" --> F{net/http: 调用 LLM API};
        F -- "HTTPS Request" --> G[外部 LLM 服务];
        G -- "HTTPS Response (JSON)" --> F;
        F -- "返回生成的 Commit Message" --> D;
        D -- "os.WriteFile" --> H[4. Git 提交信息文件 (.git/COMMIT_EDITMSG)];
        H --> I[IDE 显示已填充的提交信息];
        I --> J[开发者确认并完成提交];
    end

    subgraph "外部服务"
        G
    end
```

## 3. 核心组件设计

### 3.1. Git Hook: `prepare-commit-msg`
- **技术：** Shell 脚本 (`/bin/sh`)。
- **职责：**
    1.  作为系统的入口点，由 Git 自动调用。
    2.  接收 Git 传入的参数，最重要的是第一个参数：`COMMIT_EDITMSG` 文件的路径。
    3.  调用编译好的 Go 核心程序，并将所有接收到的 Git 参数原样传递给它。
    4.  确保 Go 程序有可执行权限。

### 3.2. 核心逻辑程序 (`gac-core`)
- **技术：** Go (Golang) 1.18+，编译为单个可执行二进制文件。
- **职责：**
    1.  **参数解析：** 解析从 Shell 脚本传来的参数，获取 `COMMIT_EDITMSG` 文件路径。
    2.  **执行 Git 命令：** 使用 Go 的 `os/exec` 标准库包执行 `git diff --staged`，并捕获其标准输出作为代码变更内容。
    3.  **配置管理：** 从本地配置文件（如 `.gac_config.json`）或环境变量中读取 LLM API Key、Endpoint 和 Prompt 模板。
    4.  **调用 LLM API：** 使用标准库 `net/http` 封装并发送 HTTP POST 请求到指定的 LLM 服务端点。使用 `encoding/json` 解析返回的 JSON 数据。
    5.  **写回文件：** 使用 `os.WriteFile` 将 LLM 返回的、经过清洗和格式化的提交信息，覆盖写入到 `COMMIT_EDITMSG` 文件中。
    6.  **异常处理：** 通过 `log` 包和 `error` 处理，确保在任何步骤失败（如网络问题、API限流）时都能静默退出，不阻塞用户的 Git 操作。

## 4. 数据流

1.  用户在 IDE 中点击 "Commit".
2.  Git 运行 `.git/hooks/prepare-commit-msg` 脚本，并传入 `.git/COMMIT_EDITMSG` 的路径。
3.  Shell 脚本执行位于 `.gac/` 目录下的 Go 程序：`./.gac/gac-core .git/COMMIT_EDITMSG`。
4.  `gac-core` 程序执行 `git diff --staged` 得到 `diffOutput`。
5.  `gac-core` 构建 Prompt, 例如: `prompt := fmt.Sprintf("根据以下代码变更生成提交信息：\n%s", diffOutput)`。
6.  `gac-core` 通过 `net/http` 发送 `prompt` 到 LLM API。
7.  LLM API 返回 JSON 响应，`gac-core` 将其解析到结构体中。
8.  `gac-core` 提取出核心的提交信息文本。
9.  `gac-core` 使用 `os.WriteFile` 打开 `.git/COMMIT_EDITMSG` 文件，将该文本写入。
10. 脚本执行完毕，IDE 加载 `.git/COMMIT_EDITMSG` 文件内容并显示。

## 5. 部署与安装

为了方便使用，项目将提供一个 `install.sh` 脚本用于 Unix-like 系统 (Linux, macOS) 和一个 `install.bat` 脚本用于 Windows。这些脚本会负责编译 Go 程序并设置 Git 钩子。

### `install.sh` (for Linux/macOS)
```bash
#!/bin/sh
# 项目内 Go 源码路径
GAC_SOURCE_DIR=".gac"
# Go 程序编译后的目标路径
GAC_TARGET_BINARY="$GAC_SOURCE_DIR/gac-core"
# Git 钩子路径
HOOK_DIR=".git/hooks"
HOOK_FILE="$HOOK_DIR/prepare-commit-msg"

# 1. 编译 Go 程序
echo "Building Go application..."
go build -o "$GAC_TARGET_BINARY" "$GAC_SOURCE_DIR/main.go"
if [ $? -ne 0 ]; then
    echo "Go build failed. Please check your Go environment."
    exit 1
fi

# 2. 确保钩子目录存在
mkdir -p "$HOOK_DIR"

# 3. 创建或覆盖 prepare-commit-msg 文件
echo "Creating Git hook..."
cat << EOF > "$HOOK_FILE"
#!/bin/sh
# 调用核心 Go 程序
$GAC_TARGET_BINARY "\$1"
EOF

# 4. 赋予可执行权限
chmod +x "$HOOK_FILE"
chmod +x "$GAC_TARGET_BINARY"

echo "AI Commit Message Generator installed successfully!"
```
开发者只需在项目根目录执行一次 `sh install.sh` 即可完成安装。
