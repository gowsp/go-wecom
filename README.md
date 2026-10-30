# go-wecom

企业微信（WeCom）API 的 Go SDK，支持应用消息、微信客服、文档、邮件、素材、群机器人及回调加解密。

- **开箱即用**：统一入口，共享 access_token 缓存，自动获取和刷新 token。
- **按需接入**：各功能独立成包，可单独使用。
- **专注接口**：只封装协议与 API，不读取配置文件，不包含业务流程或业务数据存储。

> 模块路径为 `github.com/gowsp/go-wecom`，根包名为 `wecom`。需要 Go 1.22 或更高版本。

[快速开始](#快速开始) · [功能导航](#功能导航) · [功能覆盖](#功能覆盖) · [接入方式](#接入方式) · [调用约定](#调用约定) · [更多资料](#更多资料)

## 快速开始

### 1. 安装

```bash
go get github.com/gowsp/go-wecom
```

### 2. 准备凭据

调用企业微信应用 API 通常需要以下参数：

| 参数 | 含义 |
| --- | --- |
| `CorpID` | 企业 ID |
| `CorpSecret` | 对应应用或功能的 Secret |
| `AgentID` | 应用 ID，发送应用消息时使用 |

**不同功能可能使用不同的 Secret 和权限范围。** 聚合入口不会扩大凭据权限；请按企业微信后台的配置，使用对应凭据创建实例。

### 3. 发送一条应用消息

```go
package main

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/gowsp/go-wecom"
)

func main() {
    agentID, err := strconv.Atoi(os.Getenv("WECOM_AGENT_ID"))
    if err != nil || agentID <= 0 {
        log.Fatal("请设置有效的 WECOM_AGENT_ID")
    }

    api := wecom.New(
        os.Getenv("WECOM_CORP_ID"),
        os.Getenv("WECOM_CORP_SECRET"),
        agentID,
    )

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := api.Message.SendText(ctx, "zhangsan", "你好，企业微信！"); err != nil {
        log.Fatal(err)
    }
}
```

示例由调用方读取环境变量；SDK 本身不读取环境变量。运行前设置上述三个变量，并将 `zhangsan` 替换为应用可见范围内的成员 UserID。

## 功能导航

### 应用 API

通过 `wecom.New` 一次创建，所有字段共享同一个 `client.Client` 和 token 缓存；也可单独创建子包实例。

| 功能 | 聚合入口 | 子包 | 主要能力 |
| --- | --- | --- | --- |
| 通讯录 | `api.Contact` | [contact](https://pkg.go.dev/github.com/gowsp/go-wecom/contact) | 成员、部门、标签、查看规则、异步导入导出 |
| 客户联系 | `api.External` | [external](https://pkg.go.dev/github.com/gowsp/go-wecom/external) | 客户、企业标签、客户群、联系我、继承、群发、欢迎语 |
| 应用管理 | `api.Agent` | [agent](https://pkg.go.dev/github.com/gowsp/go-wecom/agent) | 应用信息、自定义菜单、工作台 |
| 身份验证 | `api.Auth` | [auth](https://pkg.go.dev/github.com/gowsp/go-wecom/auth) | 网页授权、code 换身份、Web 登录、二次验证 |
| 应用消息 | `api.Message` | [message](https://pkg.go.dev/github.com/gowsp/go-wecom/message) | 文本、Markdown、图片、语音、视频、文件、卡片、图文、小程序通知、撤回、模板卡片更新 |
| 应用群聊 | `api.AppChat` | [appchat](https://pkg.go.dev/github.com/gowsp/go-wecom/appchat) | 创建、更新、查询群聊及发送文本 |
| 微信客服 | `api.KF` | [kf](https://pkg.go.dev/github.com/gowsp/go-wecom/kf) | 账号、接待人员、会话、消息、客户、统计、知识库 |
| 微盘 | `api.WeDrive` | [wedrive](https://pkg.go.dev/github.com/gowsp/go-wecom/wedrive) | 空间、文件、权限与分享 |
| 文档 | `api.WeDoc` | [wedoc](https://pkg.go.dev/github.com/gowsp/go-wecom/wedoc) | 文档、表格、智能表格、智能文档、收集表、权限与上传 |
| 邮件 | `api.Mail` | [mail](https://pkg.go.dev/github.com/gowsp/go-wecom/mail) | 收发邮件、文件夹、联系人、群组与公共邮箱 |
| 待办 | `api.Todo` | [todo](https://pkg.go.dev/github.com/gowsp/go-wecom/todo) | 待办详情与完成状态 |
| 公费电话 | `api.Dial` | [dial](https://pkg.go.dev/github.com/gowsp/go-wecom/dial) | 拨打记录 |
| 紧急通知 | `api.PSTN` | [pstn](https://pkg.go.dev/github.com/gowsp/go-wecom/pstn) | 语音电话与接听状态 |
| 会议与会议室 | `api.MeetingRoom` | [meetingroom](https://pkg.go.dev/github.com/gowsp/go-wecom/meetingroom) | 会议室管理与预定 |
| 电子发票 | `api.Invoice` | [invoice](https://pkg.go.dev/github.com/gowsp/go-wecom/invoice) | 查询发票、更新报销状态 |
| 会话内容存档 | `api.MsgAudit` | [msgaudit](https://pkg.go.dev/github.com/gowsp/go-wecom/msgaudit) | 存档权限、同意查询、群信息、机器人信息、RSA/AES 解密 |
| 小程序支付 | `api.MiniAppPay` | [miniapppay](https://pkg.go.dev/github.com/gowsp/go-wecom/miniapppay) | 下单、签名、退款、账单、支付/退款通知 |
| 企业支付 | 独立构造 | [payment](https://pkg.go.dev/github.com/gowsp/go-wecom/payment) | 商户证书 + XML：红包、向员工付款；用 `payment.New` 构造，不在聚合入口内 |
| 高级功能 | `api.Advanced` | [advanced](https://pkg.go.dev/github.com/gowsp/go-wecom/advanced) | 申请单查询与审批信息回写 |
| 对外收款 | `api.ExternalPay` | [externalpay](https://pkg.go.dev/github.com/gowsp/go-wecom/externalpay) | 商户、收款记录、资金流水、商户申请 |
| 安全管理 | `api.Security` | [security](https://pkg.go.dev/github.com/gowsp/go-wecom/security) | 文件/截屏记录、可信设备、操作日志、域名 |
| 上下游/企业互联 | `api.CorpGroup` | [corpgroup](https://pkg.go.dev/github.com/gowsp/go-wecom/corpgroup) | 共享应用、下游 token、链联系人、对接规则 |
| 账号 ID 转换 | `api.IDConvert` | [idconvert](https://pkg.go.dev/github.com/gowsp/go-wecom/idconvert) | open_userid/open_corpid/tmp_external_userid 转换 |
| 审批 | `api.Approval` | [approval](https://pkg.go.dev/github.com/gowsp/go-wecom/approval) | 模板、提交、查询、假期额度与回调 |
| 打卡 | `api.Checkin` | [checkin](https://pkg.go.dev/github.com/gowsp/go-wecom/checkin) | 打卡记录、排班、补卡、规则与设备数据 |
| 日程 | `api.Schedule` | [schedule](https://pkg.go.dev/github.com/gowsp/go-wecom/schedule) | 日历、日程、参与人与回执回调 |
| 直播 | `api.Live` | [live](https://pkg.go.dev/github.com/gowsp/go-wecom/live) | 预约直播、观看明细、分享码 |
| 汇报 | `api.Journal` | [journal](https://pkg.go.dev/github.com/gowsp/go-wecom/journal) | 汇报记录与统计 |
| 人事助手 | `api.HR` | [hr](https://pkg.go.dev/github.com/gowsp/go-wecom/hr) | 花名册字段与员工信息 |
| 素材 | `api.Media` | [media](https://pkg.go.dev/github.com/gowsp/go-wecom/media) | 临时素材、图片、语音及异步上传 |

### 独立接入

以下功能不使用聚合入口，也不共享应用的 access_token：

| 功能 | 子包 | 所需凭据 |
| --- | --- | --- |
| 群机器人推送 | [webhook](https://pkg.go.dev/github.com/gowsp/go-wecom/webhook) | webhook URL 中的 key |
| 回调验签与加解密 | [callback](https://pkg.go.dev/github.com/gowsp/go-wecom/callback) | Token、EncodingAESKey、CorpID |
| 智能机器人（API 模式） | [aibot](https://pkg.go.dev/github.com/gowsp/go-wecom/aibot) | 回调返回的 response_url |
| 智能机器人（长连接） | [aibot/longconn](https://pkg.go.dev/github.com/gowsp/go-wecom/aibot/longconn) | BotID、Secret |

### 功能覆盖

当前官方服务端目录中，除家校沟通、家校应用、政民沟通（按既定范围排除，清单中按“不适用”列出）外，其余接口全部已实现，清单里没有待办项：共 565 个已支持页面、38 个包。

说明性页面（概述、接入指引、签名算法等）只作为链接列出；数据与智能专区中属于“专区程序侧 SDK”的页面同样按不适用标注，因为那是运行在专区程序里的 SDK，而非企业可调用的 HTTP 接口。逐项状态、官方文档链接与实现源码位置见 [功能覆盖清单](docs/coverage.md)，接入前请核对所需接口。

## 接入方式

以下片段省略 import 和外围函数；完整示例见 [example_test.go](example_test.go)。

### 使用聚合入口

```go
api := wecom.New(corpID, corpSecret, agentID)

err := api.Message.SendText(ctx, userID, "通知内容")
```

`api.Client` 是共享传输层。无需发送应用消息时，可将 `agentID` 设为 `0`；其他功能仍需对应的 Secret 和接口权限。

### 只使用某个子包

```go
shared := client.New(corpID, corpSecret)
docs := wedoc.New(shared)

doc, err := docs.CreateDoc(ctx, wedoc.DocTypeSmartsheet, "项目跟进")
```

将同一个 `shared` 传给多个子包即可复用 token 缓存。不同凭据应使用不同的 `client.Client`。

### 自定义 HTTP 客户端

```go
api := wecom.New(corpID, corpSecret, agentID,
    client.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

默认 HTTP 超时为 **5 秒**，上传等耗时请求可按需调整。测试或自建网关可通过 `client.WithBaseURL(baseURL)` 修改 API 地址。

### 群机器人推送

```go
hook := webhook.New(key)
err := hook.Send(ctx, webhook.Message{
    MsgType: "text",
    Text:    &webhook.Text{Content: "构建完成"},
})
```

### 回调验签与解密

```go
crypto, err := callback.NewCallbackCrypto(token, encodingAESKey, corpID)
if err != nil {
    return err
}

// 首次配置回调 URL 时，验签并取得需要原样返回的明文。
echo, err := crypto.VerifyURL(msgSignature, timestamp, nonce, echostr)
```

接收消息使用 `DecryptXML` 或 `DecryptJSON`，加密被动回复使用 `EncryptReply`。HTTP 路由、响应写入、事件分发和回复内容由调用方实现。

### 智能机器人回调（API 模式）

```go
message, err := aibot.ParseMessage(plaintext)   // 先按回调配置解密
if err != nil {
    return err
}
if message.MsgType == aibot.TypeText {
    err = client.Response(ctx, message.ResponseURL, aibot.MarkdownReply("收到"))
}
```

`response_url` 由回调给出，1 小时内有效且只能用一次。`ParseEvent` 解析 `enter_chat`、`template_card_event` 等事件，未知字段保留在 `Raw` 中。

### 智能机器人长连接

同一套消息模型：`event.Message` 与 API 模式的 `aibot.Message` 是同一个类型，`reply.Reply` 接受 `aibot.Reply`。

```go
conn := longconn.New(botID, secret,
    longconn.WithHandler(func(ctx context.Context, event longconn.Event, reply *longconn.Responder) error {
        if event.Command == longconn.MessageCallback {
            return reply.Reply(ctx, aibot.TextReply("已收到"))
        }
        return nil // 不回复其他事件
    }),
)
err := conn.Start(ctx)
```

`Start` 阻塞运行，取消 `ctx` 后正常返回。默认不回复消息；由 handler 决定回复时机与内容，进入会话事件可调用 `reply.WelcomeText`。handler 串行执行，应及时返回并遵守 `ctx`，不要阻塞心跳。

连接中断自动退避重连；订阅被拒绝时返回 `*longconn.SubscriptionError`（系统繁忙 `-1` 除外）。handler 错误通过 `*longconn.HandlerError` 返回，不自动重放业务逻辑。SDK 不输出全局日志，可通过 `longconn.WithObserver` 接收连接与重试诊断。

## 调用约定

### Token 与重试

- 复用客户端实例，避免每次调用都重新创建 token 缓存。
- 应用 API 自动获取、缓存并提前刷新 access_token，无需手动传入。
- 收到 token 无效或过期错误（`40014`、`42001`）时，刷新后重试一次；其他错误不自动重试。
- 网络失败后的业务重试和幂等控制由调用方决定，避免重复发送消息或写入数据。

### 错误处理

应用 API 返回非零 `errcode` 时，可通过 `errors.As` 获取企业微信错误码：

```go
if err != nil {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        log.Printf("企业微信接口失败：操作=%s，错误码=%d", apiErr.Op, apiErr.Errcode)
    } else {
        log.Printf("调用失败：%v", err)
    }
}
```

请勿在日志中输出 Secret、access_token、回调密钥或带凭据的完整请求 URL。

### 分页与上传

- **分页**：SDK 每次只请求一页，由调用方循环拉取。以对应接口的 `NextCursor`、`HasMore` 或 `NextStart` 为准；不同接口的 `HasMore` 可能是整数、布尔值或字符串，不要混用判断条件。
- **素材上传**：使用 `api.Media.Upload`；群机器人素材使用 `hook.UploadMedia`，两者鉴权方式不同。
- **文档上传**：使用 `api.WeDoc.UploadImage`、`UploadFile`；大文件可使用 `InitChunkUpload` → `UploadChunk` → `FinishChunkUpload`。
- **微盘上传**：10 MiB 以内用 `api.WeDrive.UploadFile` 配合 `wedrive.InlineBase64`；更大的文件用 `api.WeDrive.UploadChunked`（自动完成 2 MiB 分块的 `file_upload_init` → `file_upload_part` → `file_upload_finish`，分块摘要由 `wedrive.BlockSHA` 按官方附录计算）。
- **响应大小**：应用 API 的 JSON 响应默认上限为 1 MiB，批量类接口可用 `client.WithResponseLimit(bytes)` 提高；二进制下载默认最多缓冲 32 MiB，可用 `client.WithDownloadLimit(bytes)` 调整。超限返回 `client.ErrResponseTooLarge`，不返回截断数据。
- **接口限制**：权限、频率、文件大小与格式均以对应的企业微信官方文档为准。

### 更新接口的覆盖与增量

企业微信不同领域的更新接口语义不同，SDK 按官方文档区别建模：

**日历与日程是覆盖式更新**，请求体即完整的新状态，未填的字段会被清空：

- 日历更新的 `summary`、`color` 必填；日程更新的 `start_time`、`end_time` 必填，`UpdateSchedule` 会用返回的 `schedule_id` 反映重复日程被拆分后的新 ID。
- 只想改其中一部分时，先读取再整体提交。
- 集合默认被替换；`CalendarUpdate.SkipPublicRange` 与 `EntryUpdate.SkipAttendees` 分别用于保留可订阅范围和参与人。
- `cal_id`、`is_public`、`is_corp_calendar` 创建后不可修改，因此不出现在更新结构体里。

```go
if _, err := api.Schedule.UpdateCalendar(ctx, schedule.CalendarUpdate{
    CalID:   "cal1",
    Summary: "改名",
    Color:   "#FF3030",
    Admins:  []string{"u1"},
}); err != nil {
    return err
}
```

**会议更新是增量式**，`meeting.UpdateRequest` 只发送显式设置的字段：

| 写法 | 含义 |
| --- | --- |
| 字段为 `nil` | 不修改，保持原值 |
| 字段为指向零值的指针 | 设置为零值，例如清空描述、关闭开关 |
| 切片为 `nil` | 不修改集合 |
| 切片为非 nil 空值 | 清空集合 |

`meeting_start` 与 `meeting_duration` 必须成对出现；重复会议配置放在 `meeting.Reminders` 对象里，其 `remind_before` 是提醒时长列表，不是按成员配置。

### 设计边界

SDK 管理 token 缓存、连接与心跳等技术状态，不决定业务回复、数据存储或业务重试。配置来源、HTTP 回调路由和业务编排由应用负责。优先沿用官方 API 领域与术语，提供必要的 Go 便利方法，不强行统一不同接口的语义。

## 更多资料

- [API 参考](https://pkg.go.dev/github.com/gowsp/go-wecom)：查看全部公开类型、方法与参数。
- [示例代码](example_test.go)：应用消息、邮件、智能表格、素材、群机器人及回调等示例。
- [企业微信开发者中心](https://developer.work.weixin.qq.com/)：接口权限与接入要求；SDK 的 API 方法注释也附有对应官方文档链接。

当前尚未发布 `v1`，API 可能发生不兼容调整；生产接入建议固定具体版本或提交。

## 开发与维护

### 本地验证

```bash
go test ./...
go vet ./...
go build ./...
```

普通 `go test` 离线、只读运行，不抓取官网或改写文件；覆盖清单测试会校验清单与代码一致，包括实现源码的位置。

### 更新覆盖清单

[功能覆盖清单](docs/coverage.md) 由 [coverage_test.go](coverage_test.go) 生成，请勿手工编辑。是否支持由 [能力台账](docs/capabilities.json) 决定，源码链接由方法注释扫描得出；台账、源码与清单三者不一致时测试失败。在仓库根目录按需运行：

```bash
# 在线：重新抓取官方目录树，更新条目、支持状态与源码链接
go test -run TestCoverageDocument -update-coverage .

# 离线：沿用清单中已有的目录树，仅刷新源码链接，不增删官方条目
go test -run TestCoverageDocument -update-coverage-sources .
```

### 补充未覆盖接口

1. 按企业微信 API 领域组织子包（例如 `contact`、`meeting`），复用 `client.Client`，避免重复实现传输与 token 管理。
2. 方法注释末尾附官方文档地址：`// 文档: https://developer.work.weixin.qq.com/document/path/NNNNN`。
3. 用 `httptest` 配合 `client.WithBaseURL` 编写请求与错误分支测试。
4. 在 [能力台账](docs/capabilities.json) 中登记该页面：`status` 写 `implemented`，只实现了一部分写 `partial` 并在 `notes` 说明缺口；未取得线上验证时 `integration` 保持 `untested`。
5. 重新生成覆盖清单：需要新增或移除官方条目时使用在线命令；仅更新已有条目的源码链接时使用离线命令，然后运行本地验证。

## 许可证

[Apache License 2.0](LICENSE)
