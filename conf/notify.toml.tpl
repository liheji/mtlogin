# 全局通知超时时间，单位秒，支持小数；必须 > 0。
timeout = 10

# [qqpush]
# enabled: 是否启用 QQPUSH 通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理。
# qq: 接收推送的 QQ 号，对应旧环境变量 QQPUSH。
# token: QQPUSH 推送 token，对应旧环境变量 QQPUSH_TOKEN。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[qqpush]
enabled = false
use_system_proxy = false
qq = ""
token = ""
timeout = 0

# [weixin]
# enabled: 是否启用企业微信通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理。
# corp_id: 企业微信企业 ID，对应旧环境变量 WXCORPID。
# agent_secret: 企业微信应用密钥，对应旧环境变量 WXAGENTSECRET。
# agent_id: 企业微信应用 ID，对应旧环境变量 WXAGENTID。
# user_id: 接收消息的成员 ID，多个成员用 | 分隔；@all 表示发送给所有成员。对应旧环境变量 WXUSERID。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[weixin]
enabled = false
use_system_proxy = false
corp_id = ""
agent_secret = ""
agent_id = 0
user_id = "@all"
timeout = 0

# [dingtalk]
# enabled: 是否启用钉钉机器人通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理。
# webhook_token: 钉钉机器人 webhook 地址中的 access_token，对应旧环境变量 DING_TALK_ROBOT_WEBHOOK_TOKEN。
# secret: 钉钉机器人安全设置中的加签密钥，对应旧环境变量 DING_TALK_ROBOT_SECRET。
# at_mobiles: 需要 @ 的手机号列表；留空（[]）表示 @ 所有人。对应旧环境变量 DING_TALK_ROBOT_AT_MOBILES。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[dingtalk]
enabled = false
use_system_proxy = false
webhook_token = ""
secret = ""
at_mobiles = []
timeout = 0

# [telegram]
# enabled: 是否启用 Telegram Bot 通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理；如果 proxy 非空，则优先使用 proxy。
# bot_token: Telegram Bot token，需要在 BotFather 创建机器人后获取。对应旧环境变量 TGBOT_TOKEN。
# chat_id: Telegram 聊天 ID，可通过 getUpdates 接口查看 chat.id。对应旧环境变量 TGBOT_CHAT_ID。
# proxy: Telegram Bot API 专用代理地址；支持 http://host:port、http://user:pass@host:port、socks5://host:port。对应旧环境变量 TGBOT_PROXY。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[telegram]
enabled = false
use_system_proxy = false
bot_token = ""
chat_id = 0
proxy = ""
timeout = 0

# [feishu]
# enabled: 是否启用飞书机器人通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理。
# webhook_url: 飞书机器人 Webhook URL，对应旧环境变量 FEISHU_WEBHOOKURL。
# secret: 飞书机器人安全设置中的签名密钥；不使用签名时留空。对应旧环境变量 FEISHU_SECRET。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[feishu]
enabled = false
use_system_proxy = false
webhook_url = ""
secret = ""
timeout = 0

# [ntfy]
# enabled: 是否启用 Ntfy 通知。
# use_system_proxy: 是否使用 app.toml 中 [system].proxy 作为请求代理。
# url: Ntfy 服务地址，例如 https://ntfy.sh 或自建服务地址。对应旧环境变量 NTFY_URL。
# topic: Ntfy 主题，对应旧环境变量 NTFY_TOPIC。
# username: Ntfy Basic Auth 用户名，需要和 password 一起设置。对应旧环境变量 NTFY_USER。
# password: Ntfy Basic Auth 密码，需要和 username 一起设置。对应旧环境变量 NTFY_PASSWORD。
# token: Ntfy Bearer token；与 username/password 作用一致，且优先级更高。对应旧环境变量 NTFY_TOKEN。
# timeout: 当前通道超时时间，单位秒，支持小数；0 表示继承全局 timeout。
[ntfy]
enabled = false
use_system_proxy = false
url = ""
topic = ""
username = ""
password = ""
token = ""
timeout = 0
