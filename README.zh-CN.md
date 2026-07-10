# Git Worktree Manager

`gwt` 是一个 Git worktree 生命周期管理工具，提供适合人工操作的 TUI，以及适合脚本和智能体调用的 CLI/JSON 输出。

## 安装

### 从源码安装

```bash
go install github.com/kinjaze/git-worktree-manager/cmd/gwt@latest
```

### 从 Release 压缩包安装

下载适合你平台的压缩包，解压后把 `gwt` 放到 `PATH` 中：

```bash
tar -xzf gwt_<version>_<os>_<arch>.tar.gz
chmod +x gwt
sudo mv gwt /usr/local/bin/
gwt --help
```

Windows 用户下载 `.zip` 包，解压后将 `gwt.exe` 所在目录加入 `PATH` 即可。

## 核心命令

```bash
gwt create login-page --repo /path/to/repo --source origin/main --branch feat/login-page --path /path/to/worktree
gwt list
gwt update <id-or-path>
gwt merge-back <id-or-path>
gwt remove <id-or-path>
gwt tui
```

## 合并行为

- `gwt update` 会从记录的源分支执行普通 merge。
- `gwt merge-back` 会执行 `git merge --no-ff <worktreeBranch>`。
- 如果发生冲突，工具会保留 Git 原生冲突状态，并给出后续处理步骤。
- 如果目标 worktree 存在未提交变更，`merge-back` 会停止，避免覆盖本地工作。

## 语言

默认语言是英文，可以启用中文：

```bash
gwt --lang zh <command>
gwt config set language zh
```

JSON 字段名、状态值和错误码保持英文，便于脚本稳定解析。

## 常用场景

### 启动交互式界面

```bash
gwt tui
```

### 创建受管理的 worktree

```bash
gwt create feature-a \
  --repo /path/to/repo \
  --source origin/main \
  --branch feat/feature-a \
  --path /path/to/feature-a
```

### 输出 JSON

```bash
gwt list --json
```

## 校验安装包

Release 中包含 `checksums.txt`，可以用 SHA-256 校验下载文件：

```bash
shasum -a 256 -c checksums.txt
```
