# 用户故事：实现核心提交信息生成逻辑

**作为一名开发者,** 我希望能构建一个名为 `gac-core` 的 Go 应用程序。这个程序需要能够：
1.  执行 `git diff --staged` 来获取暂存的代码变更。
2.  将获取到的代码 `diff` 内容通过 HTTP 请求发送给一个大语言模型 (LLM) API。
3.  接收并解析 LLM 的响应，提取出生成的提交信息。
4.  将最终的提交信息输出，以便 Git 钩子脚本可以使用它。

**验收标准:**
- `gac-core` 应用程序可以被成功编译和执行。
- 当在有暂存变更的 Git 仓库中运行时，程序能够捕获 `git diff --staged` 的输出。
- 程序能够成功地向指定的 LLM API 端点发送 `diff` 内容。
- 程序能够正确解析 LLM API 的 JSON 响应。
- 程序最终将从 LLM 获取并处理过的提交信息打印到标准输出。
- **输出的提交信息必须符合 Conventional Commits 规范。**
    - **规范描述:** 提交信息格式应为 `<type>(<scope>): <description>`，其中 `type` 主要是 `feat` 或 `fix`。
    - **提交信息示例:**
      ```
      feat(parser): add ability to parse arrays

      - Adds a new function `parseArray` to handle array literals.
      - Integrates `parseArray` into the main parsing loop.

      Fixes: #42
      ```
- 如果过程中任何步骤失败，程序应以非零状态码退出，并输出有意义的错误信息到标准错误流。
