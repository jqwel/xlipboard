<div align="center">
  <img src="https://raw.githubusercontent.com/jqwel/xlipboard/main/src/static/design64.png" style="display: inline-block; vertical-align: middle;">

[//]: # (  <img src="https://github.com/jqwel/xlipboard/blob/229d323ff817118fe0aff0baca0f9c38e5f63f42/src/static/design64.png?raw=true" style="display: inline-block; vertical-align: middle;">)
  <h1 style="display: inline-block; vertical-align: middle;">xlipboard</h1>
</div>

![GitHub release (latest by date)](https://img.shields.io/github/v/release/jqwel/xlipboard?logo=go)

xlipboard 是一款可以帮你在Windows、Ubuntu和macOS桌面系统之间同步剪切板文件的应用

## 文档
【[中文](https://github.com/jqwel/xlipboard/blob/master/README.md)】

## 安装
【[必备条件](https://github.com/jqwel/xlipboard/tree/main/prerequisite)】

## 配置

`xlipboard.exe` 将在运行路径下面创建两个文件： `Config.json` ~~和 `log.txt`~~

你可以通过修改 `Config.json` 来自定义配置

### `Config.json`

- `Port` # 服务端口
  - 类型: `string`
  - 默认: `"3216"`

- `Authkey` # 认证密钥
  - type: `string`
  - default: `随机生成不少于64位`

- `NtpAddress` # 网络时间协议同步时间
  - type: `string`
  - default: `ntp.aliyun.com`

- `SyncSettings` # 远程同步其他服务器的地址和端口
  - type: `object`
  - children:
    - `Target`
      - type: `string`
      - default: `"192.168.200.101:3216"`
