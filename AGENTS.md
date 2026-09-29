# AGENTS.md

## 项目定位与边界

Linux / macOS 系统信息 CLI，基于 zcalusic/sysinfo 增强。直接执行输出 JSON，辅助判断二进制构建与运行环境的兼容性；不自动作兼容性结论。

## 代码地图

```text
cmd/sysinfo/       # CLI 入口，输出一次 JSON
*.go              # 沿用上游系统信息采集结构
runtime*.go       # 基础页大小与平台运行环境
platform_*.go     # Linux / macOS 采集入口
libc.go           # 静态读取已安装 libc 的版本证据
cpuid/            # 上游 CPU 指令封装
```

## 关键约定

- 采集只读，不创建或修复系统文件，不执行外部程序。
- 标准库实现；发布使用 CGO_ENABLED=0，Linux 静态链接，macOS 仅使用系统自带库，不依赖额外语言运行时。
- 缺失证据不转换为正常默认值；新增运行环境采集错误保存在 runtime.errors。
- 保留上游版权及许可证。版本由 version.go 管理。
- 验证入口：make fix、make lint、make test、make build；build 一次产出全部支持架构，跨编译不等于客户机运行验证。

## References

- [README.md](README.md)：使用、构建和采集边界。
