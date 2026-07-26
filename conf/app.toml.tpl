[system]
# 系统代理地址，供显式开启 use_system_proxy 的通知通道使用。
# 支持 http://host:port、http://user:pass@host:port、socks5://host:port；不使用代理时留空。
proxy = ""

[app]
# 定时任务表达式，使用 robfig/cron v3 格式。
# 支持 5 字段（分 时 日 月 周）或 6 字段（秒 分 时 日 月 周，秒可选）。
# 也支持描述符，如 @every 1h30m、@daily、@hourly。
crontab = "2 */2 * * *"
# 随机延迟范围，单位秒，支持小数；min_delay = 0 且 max_delay = 0 表示立即执行。
# 实际延迟在 [min_delay, max_delay] 范围内均匀随机。
min_delay = 0
max_delay = 0

[refresh]
# 单次刷新业务请求总超时，单位秒，支持小数；必须 > 0。
timeout = 300
# 账号密码模式认证失败后是否在本次刷新内立即重新登录。
# false：清空本地身份后等下次刷新再登录；true：清空后立即重新登录并继续本次刷新。
retry_on_auth_failure = false

[mt]
# M-Team API 专用代理地址。
# 支持 http://host:port、http://user:pass@host:port、socks5://host:port；不使用代理时留空。
proxy = ""
# M-Team API 域名，默认 api.m-team.io。必填。
api_host = "api.m-team.io"
# M-Team API 请求 Referer 头。
referer = "https://kp.m-team.cc/"
# M-Team API 请求超时，单位秒，支持小数；必须 > 0。
timeout = 60

# 账号密码认证：填写 username/password，可选 totp_secret。
# 程序会优先读取 data/cookie.db 中的本地身份，缺少时使用此处的用户名和密码发起登录。
[mt.password_auth]
username = ""
password = ""
# TOTP 两步验证密钥（Base32）；未开启两步验证时留空。
totp_secret = ""

# Token 认证：填写浏览器抓到的 Authorization、DID 和 visitorid。
# 三项只从配置读取，程序不会自动更新；过期后需要重新抓取并重启。
# 至少配置 password_auth 或 token_auth 一种；两种都配置时 Token Auth 优先。
[mt.token_auth]
# 浏览器请求头中的 Authorization 值。
auth = ""
# 浏览器请求头中的 Did 值。
did = ""
# 浏览器请求头中的 visitorid 值。
visitor_id = ""

# 请求头参数：发送到 M-Team API 的固定请求头值。
[mt.header]
# 请求头 version 的值，对应 M-Team App 版本号。
version = "1.1.4"
# 请求头 webversion 的值，对应 M-Team Web 前端构建版本号。
web_version = "1140"
# 请求头 User-Agent 的值；同时用于选择最接近的 tls-client 内置浏览器 profile。
# 项目不支持、也不会读取外部配置的 JA3 或 JA4。
user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0"
