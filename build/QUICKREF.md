# Wails v3 快速参考

## 开发命令

```bash
# 开发模式（热重载）
wails3 dev
# 或
task dev

# 构建当前平台
task build

# 打包（创建安装包）
task package

# 运行构建的应用
task run
```

## 平台特定构建

```bash
# Windows
task windows:build
task windows:package  # 创建 NSIS 安装程序

# Linux
task linux:build
task linux:package    # 创建 AppImage + tar.gz

# macOS
task darwin:build
task darwin:package   # 创建 DMG
```

## 跨平台编译

```bash
# 一次性设置 Docker（约 800MB 下载）
task setup:docker

# 然后可以从任何平台构建
task windows:build CGO_ENABLED=1
task linux:build CGO_ENABLED=1
```

## 架构指定

```bash
# 指定架构
task windows:build ARCH=amd64
task windows:build ARCH=arm64
task linux:build ARCH=arm64
task darwin:build ARCH=universal
```

## 前端任务

```bash
# 安装前端依赖
task common:install:frontend:deps

# 构建前端
task common:build:frontend

# 运行前端开发服务器
task common:dev:frontend
```

## 实用工具

```bash
# 从 appicon.png 生成平台图标
task common:generate:icons

# 生成 TypeScript 绑定
task common:generate:bindings

# 更新构建资源
task common:update:build-assets
```

## 配置文件

- `build/config.yml` - 主配置文件
- `build/windows/Taskfile.yml` - Windows 构建配置
- `build/linux/Taskfile.yml` - Linux 构建配置
- `build/darwin/Taskfile.yml` - macOS 构建配置
- `Taskfile.yml` - 主任务文件

## 环境变量

```bash
# 启用 CGO（必需）
export CGO_ENABLED=1

# 指定架构
export ARCH=amd64

# 自定义 Vite 端口
export WAILS_VITE_PORT=5173
```

## GitHub Actions

### 发布工作流

1. 进入 Actions → Release
2. 点击 "Run workflow"
3. 输入版本号（如 `v1.2.21`）
4. 点击 "Run workflow"

自动构建所有平台：Windows、Linux、macOS（AMD64 + ARM64）

### 测试构建工作流

在推送到 main 分支或 PR 时自动触发，验证所有平台能否成功构建。

## 移动平台（实验性）

```bash
# iOS（需要 macOS + Xcode）
task ios:build

# Android（需要 Android SDK）
task android:build
```

**注意**: 移动平台支持仍在 Wails v3 alpha 中，可能不稳定。

## 签名

### Windows

```yaml
# build/windows/Taskfile.yml
vars:
  SIGN_CERTIFICATE: "path/to/cert.pfx"
```

### macOS

```yaml
# build/darwin/Taskfile.yml
vars:
  SIGN_IDENTITY: "Developer ID Application: ..."
```

### Linux

```yaml
# build/linux/Taskfile.yml
vars:
  PGP_KEY: "path/to/key.asc"
```

然后运行：

```bash
wails3 setup signing  # 安全存储密码
task windows:sign     # 签名
```

## 常见问题

### CGO is disabled

```bash
export CGO_ENABLED=1
task build
```

### Task not found

```bash
# 安装 Task
go install github.com/go-task/task/v3/cmd/task@latest
```

### Linux 依赖缺失

```bash
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev libsoup-3.0-dev gcc pkg-config
```

## 完整文档

详见 `build/README.md` 和 `docs/BUILD_REQUIREMENTS.md`
