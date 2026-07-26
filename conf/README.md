# 配置文件指南

`mtlogin` 使用两个固定配置文件：

| 文件 | 作用 | 加载失败后的行为 |
| --- | --- | --- |
| `conf/app.toml` | 调度、刷新、代理和 M-Team 认证 | 程序启动失败 |
| `conf/notify.toml` | 外部通知通道 | 记录 Warning，仅启用 base 通知 |

所有时间配置均以秒为单位，并支持小数。例如，`timeout = 0.5` 表示 0.5 秒。

## 快速导航

- [初始化配置](#初始化配置)
- [选择认证方式](#选择认证方式)
- [`app.toml` 配置](#apptoml-配置)
- [`notify.toml` 配置](#notifytoml-配置)
- [代理规则](#代理规则)
- [配置验证](#配置验证)
- [常见配置问题](#常见配置问题)
- [敏感信息保护](#敏感信息保护)

## 初始化配置

从模板创建实际配置：

```bash
cp conf/app.toml.tpl conf/app.toml
cp conf/notify.toml.tpl conf/notify.toml
```

真实配置文件不会提交到 Git，也不会打入 Docker 镜像。程序只在启动时加载配置；修改后必须重启。

建议首次配置后执行：

```bash
LOCAL_TEST_RUN=true ./mtlogin
```

单次运行结束后，检查 `logs/app.log`、`logs/app.logging.wf` 和 `logs/notify.log`。

## 选择认证方式

### 认证方式对比

| 对比项 | 账号密码模式 | Token Auth 模式 |
| --- | --- | --- |
| 配置段 | `[mt.password_auth]` | `[mt.token_auth]` |
| 必填项 | `username`、`password` | `auth`、`did`、`visitor_id` |
| TOTP | 可配置 `totp_secret` | 不适用 |
| 身份持久化 | 使用 `data/cookie.db` | 不写入 LevelDB |
| 自动恢复 | 本地身份缺失时自动登录 | 需要手动更新配置 |
| 过期处理 | 清理本地身份，可选择立即重登 | 每次失败通知，不回退到密码模式 |

至少配置一种认证方式。两种方式同时配置时，Token Auth 优先。

### 账号密码模式：最小修改

在 `app.toml.tpl` 的默认配置基础上，只需填写：

```toml
[mt.password_auth]
username = "your_username"
password = "your_password"
totp_secret = ""

[mt.token_auth]
auth = ""
did = ""
visitor_id = ""
```

如果账号启用了两步验证，将 Base32 格式的 TOTP 密钥写入 `totp_secret`。

### Token Auth 模式：最小修改

从浏览器请求头获取 `Authorization`、`Did` 和 `visitorid`，并填写：

```toml
[mt.password_auth]
username = ""
password = ""
totp_secret = ""

[mt.token_auth]
auth = "your_authorization"
did = "your_did"
visitor_id = "your_visitor_id"
```

这 3 项必须同时填写。程序不会自动更新 Token Auth 配置；过期后需要重新抓取并重启。

## 配置加载规则

- 配置路径固定，不支持通过环境变量替换。
- `app.toml` 和 `notify.toml` 都会拒绝未知字段，拼写错误不会被静默忽略。
- 已删除的旧字段会返回明确错误。
- `mt.external_auth` 已重命名为 `mt.token_auth`。
- 配置只加载一次，运行期间不监听文件变化。
- `LOCAL_TEST_RUN=true` 是进程环境变量，不属于 TOML 配置。

## `app.toml` 配置

### 配置段总览

| 配置段 | 作用 |
| --- | --- |
| `[system]` | 通知通道可选使用的系统代理 |
| `[app]` | Cron 调度与随机延迟 |
| `[refresh]` | 单次刷新总超时与认证失败重试 |
| `[mt]` | M-Team API 连接参数 |
| `[mt.password_auth]` | 账号密码与 TOTP |
| `[mt.token_auth]` | 浏览器 Token Auth 身份 |
| `[mt.header]` | M-Team 固定请求头 |

### `[system]`：系统配置

```toml
[system]
proxy = ""
```

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `proxy` | string | `""` | 通知通道的共享代理。通道必须显式设置 `use_system_proxy = true` 才会使用 |

支持以下代理格式：

- `http://host:port`
- `http://user:pass@host:port`
- `socks5://host:port`

### `[app]`：调度配置

```toml
[app]
crontab = "2 */2 * * *"
min_delay = 0
max_delay = 0
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `crontab` | string | 必须是有效表达式 | 使用 robfig/cron v3，支持可选秒字段和描述符 |
| `min_delay` | float64 | `>= 0` | Cron 触发后的最小随机延迟 |
| `max_delay` | float64 | `>= min_delay` | Cron 触发后的最大随机延迟 |

随机延迟在 `[min_delay, max_delay]` 区间内均匀采样。两项都为 `0` 时立即执行。

常用 Cron 示例：

| 表达式 | 含义 |
| --- | --- |
| `2 */2 * * *` | 每 2 小时的第 2 分钟执行 |
| `0 2 */2 * * *` | 带秒字段：每 2 小时的第 2 分钟第 0 秒执行 |
| `@every 1h30m` | 每 1 小时 30 分钟执行 |
| `@daily` | 每天执行一次 |

### `[refresh]`：刷新配置

```toml
[refresh]
timeout = 300
retry_on_auth_failure = false
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `timeout` | float64 | `> 0` | 单次刷新所有业务请求的总超时 |
| `retry_on_auth_failure` | bool | 无 | 账号密码模式认证失败后，是否在本次运行中立即重新登录 |

`retry_on_auth_failure = false` 时，程序清理本地身份并等待下次调度重新登录。设置为 `true` 时，本次运行会立即重新登录并重试一次，不会无限循环。

### `[mt]`：M-Team API 配置

```toml
[mt]
proxy = ""
api_host = "api.m-team.io"
referer = "https://kp.m-team.cc/"
timeout = 60
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `proxy` | string | 可为空 | M-Team API 专用代理，不影响通知通道 |
| `api_host` | string | 必填 | API 域名，不包含 `https://` |
| `referer` | string | 建议保留默认值 | M-Team 请求的 Referer 和 Origin 来源 |
| `timeout` | float64 | `> 0` | 单个 M-Team HTTP 请求超时 |

`refresh.timeout` 控制整次刷新，`mt.timeout` 控制单个请求。整次刷新可能包含多轮 Warmup，因此通常应保证 `refresh.timeout` 大于 `mt.timeout`。

### `[mt.password_auth]`：账号密码认证

```toml
[mt.password_auth]
username = ""
password = ""
totp_secret = ""
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 与 `password` 同时填写 | M-Team 账号 |
| `password` | string | 与 `username` 同时填写 | M-Team 密码 |
| `totp_secret` | string | 可为空 | Base32 格式的 TOTP 两步验证密钥 |

账号密码模式会优先读取 `data/cookie.db`。只有本地 token 或 DID 缺失时，程序才会调用登录接口。

### `[mt.token_auth]`：Token Auth 认证

```toml
[mt.token_auth]
auth = ""
did = ""
visitor_id = ""
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `auth` | string | 3 项同时填写 | 浏览器请求头中的 `Authorization` 值 |
| `did` | string | 3 项同时填写 | 浏览器请求头中的 `Did` 值 |
| `visitor_id` | string | 3 项同时填写 | 浏览器请求头中的 `visitorid` 值 |

Token Auth 身份只从配置读取，不会写入 `data/cookie.db`，也不会被响应头自动持久化。

### `[mt.header]`：请求头参数

```toml
[mt.header]
version = "1.1.4"
web_version = "1140"
user_agent = "Mozilla/5.0 ..."
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `version` | string | 请求头 `version`，对应 M-Team App 版本 |
| `web_version` | string | 请求头 `webversion`，对应 Web 前端构建版本 |
| `user_agent` | string | 请求头 `User-Agent`，同时用于选择最接近的 tls-client 浏览器 profile |

项目使用 tls-client 内置浏览器 profile，不读取外部 JA3 或 JA4 配置。

## `notify.toml` 配置

通知 Manager 始终启用 base 通知，并将消息写入 `logs/notify.log`。所有外部通道会并发发送；单个通道失败不会阻断其他通道。

### 全局配置

```toml
timeout = 10
```

| 字段 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `timeout` | float64 | `> 0` | 所有通知通道的默认超时 |

### 通道公共字段

除 base 外，每个外部通道都有以下公共字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `enabled` | bool | 是否启用当前通道 |
| `use_system_proxy` | bool | 是否使用 `app.toml` 中的 `[system].proxy` |
| `timeout` | float64 | 当前通道超时；`0` 表示继承全局值，不能小于 `0` |

### 通道索引

| 配置段 | 通道 | 启用时的代码校验必填项 | 独立代理 |
| --- | --- | --- | --- |
| `[qqpush]` | QQPUSH | `qq`、`token` | 无 |
| `[weixin]` | 企业微信 | `corp_id`、`agent_secret` | 无 |
| `[dingtalk]` | 钉钉机器人 | `webhook_token` | 无 |
| `[telegram]` | Telegram Bot | `bot_token`、`chat_id` | `proxy` |
| `[feishu]` | 飞书机器人 | `webhook_url` | 无 |
| `[ntfy]` | Ntfy | `url`、`topic` | 无 |

### `[qqpush]`：QQPUSH

```toml
[qqpush]
enabled = false
use_system_proxy = false
qq = ""
token = ""
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `qq` | string | 接收推送的 QQ 号 |
| `token` | string | QQPUSH 推送 token |

### `[weixin]`：企业微信

```toml
[weixin]
enabled = false
use_system_proxy = false
corp_id = ""
agent_secret = ""
agent_id = 0
user_id = "@all"
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `corp_id` | string | 企业 ID |
| `agent_secret` | string | 应用密钥 |
| `agent_id` | int | 应用 ID |
| `user_id` | string | 接收成员 ID，多个成员用 `\|` 分隔；`@all` 表示全部成员 |

### `[dingtalk]`：钉钉机器人

```toml
[dingtalk]
enabled = false
use_system_proxy = false
webhook_token = ""
secret = ""
at_mobiles = []
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `webhook_token` | string | 机器人 Webhook 中的 `access_token` |
| `secret` | string | 机器人安全设置中的加签密钥，可为空 |
| `at_mobiles` | []string | 需要 @ 的手机号；空数组表示 @ 所有人 |

### `[telegram]`：Telegram Bot

```toml
[telegram]
enabled = false
use_system_proxy = false
bot_token = ""
chat_id = 0
proxy = ""
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `bot_token` | string | 通过 [BotFather](https://t.me/BotFather) 创建机器人后获取 |
| `chat_id` | int64 | 与 Bot 对话后，通过 `getUpdates` 查询 `chat.id` |
| `proxy` | string | Telegram 专用代理；非空时优先于系统代理 |

查询 `chat_id` 时，可访问：

```text
https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates
```

### `[feishu]`：飞书机器人

```toml
[feishu]
enabled = false
use_system_proxy = false
webhook_url = ""
secret = ""
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `webhook_url` | string | 飞书机器人 Webhook URL |
| `secret` | string | 签名密钥；不启用签名时留空 |

### `[ntfy]`：Ntfy

```toml
[ntfy]
enabled = false
use_system_proxy = false
url = ""
topic = ""
username = ""
password = ""
token = ""
timeout = 0
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `url` | string | Ntfy 服务地址，例如 `https://ntfy.sh` |
| `topic` | string | Ntfy 主题 |
| `username` | string | Basic Auth 用户名，需要与 `password` 一起使用 |
| `password` | string | Basic Auth 密码 |
| `token` | string | Bearer token，优先级高于 Basic Auth |

## 代理规则

| 请求类型 | 代理来源 | 优先级 |
| --- | --- | --- |
| M-Team API | `[mt].proxy` | 独立使用，不读取系统代理 |
| 普通通知通道 | `[system].proxy` | 仅在 `use_system_proxy = true` 时使用 |
| Telegram | `[telegram].proxy` 或系统代理 | 独立 `proxy` 非空时优先 |

通知代理通过请求 Context 传递。M-Team 和通知请求复用 TLS client，但 Cookie scope 相互隔离。

## 配置验证

### 启动前检查清单

- [ ] 已从 `.tpl` 创建实际 TOML 文件。
- [ ] `app.crontab` 可以被 robfig/cron v3 解析。
- [ ] `min_delay` 和 `max_delay` 均不小于 `0`，且最小值不大于最大值。
- [ ] `refresh.timeout > 0`，`mt.timeout > 0`。
- [ ] 至少完整配置一种认证方式。
- [ ] Token Auth 的 3 个字段没有只填写一部分。
- [ ] `notify.toml` 的全局 `timeout > 0`。
- [ ] 已启用通知通道的必填字段完整。
- [ ] 修改配置后已重启程序。

### 单次运行验证

```bash
LOCAL_TEST_RUN=true ./mtlogin
```

如果运行失败，先查看：

```bash
tail -n 100 logs/app.logging.wf
tail -n 100 logs/app.log
```

下游请求详情位于 `logs/http.log`，最终通知结果位于 `logs/notify.log`。

## 常见配置问题

| 错误或现象 | 原因 | 处理方式 |
| --- | --- | --- |
| `读取配置失败` | 固定路径下没有实际 TOML 文件 | 从模板复制并确认运行目录 |
| `解析配置失败` | TOML 引号、数组或段落语法错误 | 对照模板修正语法 |
| `unknown TOML key(s)` | 字段拼写错误或使用了不支持的字段 | 删除或改为当前字段名 |
| `mt.external_auth has been renamed` | 使用了旧配置段 | 改为 `[mt.token_auth]` |
| `either mt.password_auth... or mt.token_auth... is required` | 没有完整认证配置 | 完整填写一种认证方式 |
| 通知配置报错但程序继续运行 | `notify.toml` 无效，应用降级到 base 通知 | 检查 `logs/app.logging.wf` 并修正通知配置 |
| 配置修改后行为不变 | 进程仍使用启动时快照 | 重启程序或容器 |
| Docker 中找不到配置 | 空目录挂载覆盖了镜像内模板 | 在宿主机挂载目录中创建实际配置 |

## 敏感信息保护

- 不要提交 `conf/app.toml`、`conf/notify.toml` 或 `data/cookie.db`。
- 建议仅允许运行用户读取配置：`chmod 600 conf/app.toml conf/notify.toml`。
- 分享日志前仍应人工检查账号名、业务数据和论坛内容。
- 不要在 Issue、聊天记录或截图中暴露密码、Token、DID、visitor ID、Webhook 或 TOTP 密钥。
- 即使 HTTP 日志会自动脱敏，也应限制 `logs/` 目录访问权限。

