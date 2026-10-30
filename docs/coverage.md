# 功能覆盖

本文件按[企业微信官方文档目录](https://developer.work.weixin.qq.com/document)组织，分类、子分类与接口名称均取自官方侧边栏，逐项标注 SDK 的实现情况。

每个已实现的方法注释都附有官方文档地址，可与本文件交叉核对；已实现的条目还标注了实现该接口的源码位置，可直接跳转。

本文件由 `coverage_test.go` 抓取官网目录后生成，请勿手工编辑；联网刷新目录：`go test -run TestCoverageDocument -update-coverage .`；离线刷新源码链接：`go test -run TestCoverageDocument -update-coverage-sources .`。

支持状态来自 `docs/capabilities.json` 能力台账，源码位置由扫描方法注释得出，两者与本文不一致时测试失败。

## 标记说明

| 标记 | 含义 |
| --- | --- |
| `[x]` | SDK 已提供对应能力；具体支持范围以源码中的方法和类型为准 |
| `[ ]` | 尚未标记为已支持 |
| 无勾选 | 说明性文档 |
| **部分支持** | 子项只实现了一部分 |

## 快速导航

[开发指南](#开发指南) · [基础](#基础) · [连接微信](#连接微信) · [办公](#办公)

常用能力：[消息接收与发送](#消息接收与发送) · [素材管理](#素材管理) · [微信客服](#微信客服) · [邮件](#邮件) · [文档](#文档)

## 开发指南

- [开发前必读](https://developer.work.weixin.qq.com/document/path/90664)
- [基本概念介绍](https://developer.work.weixin.qq.com/document/path/90665)
- [x] [获取access_token](https://developer.work.weixin.qq.com/document/path/91039) — [client/client.go:163](../client/client.go#L163)
- [回调配置](https://developer.work.weixin.qq.com/document/path/90930)
- [x] [获取企业微信接口IP段](https://developer.work.weixin.qq.com/document/path/92520) — [iprange.go:21](../iprange.go#L21)
- [x] [获取企业微信回调IP段](https://developer.work.weixin.qq.com/document/path/92521) — [iprange.go:29](../iprange.go#L29)

## 基础

### 账号ID

- [概述](https://developer.work.weixin.qq.com/document/path/98728)
- [x] [自建应用与第三方应用的对接](https://developer.work.weixin.qq.com/document/path/95884) — [idconvert/idconvert.go:33](../idconvert/idconvert.go#L33)
- [x] [自建应用与智能机器人的对接](https://developer.work.weixin.qq.com/document/path/101521) — [idconvert/idconvert.go:33](../idconvert/idconvert.go#L33)
- [x] [tmp_external_userid的转换](https://developer.work.weixin.qq.com/document/path/98729) — [idconvert/idconvert.go:112](../idconvert/idconvert.go#L112)

### 通讯录管理

- [概述](https://developer.work.weixin.qq.com/document/path/90193)
- [x] 成员管理
  - [x] [成员扩展属性](https://developer.work.weixin.qq.com/document/path/100067) — [contact/member.go:76](../contact/member.go#L76)
  - [x] [创建成员](https://developer.work.weixin.qq.com/document/path/90195) — [contact/member.go:109](../contact/member.go#L109)
  - [x] [读取成员](https://developer.work.weixin.qq.com/document/path/90196) — [contact/member.go:125](../contact/member.go#L125)
  - [x] [更新成员](https://developer.work.weixin.qq.com/document/path/90197) — [contact/member.go:140](../contact/member.go#L140)
  - [x] [删除成员](https://developer.work.weixin.qq.com/document/path/90198) — [contact/member.go:150](../contact/member.go#L150)
  - [x] [批量删除成员](https://developer.work.weixin.qq.com/document/path/90199) — [contact/member.go:162](../contact/member.go#L162)
  - [x] [获取部门成员](https://developer.work.weixin.qq.com/document/path/90200) — [contact/member.go:178](../contact/member.go#L178)
  - [x] [获取部门成员详情](https://developer.work.weixin.qq.com/document/path/90201) — [contact/member.go:194](../contact/member.go#L194)
  - [x] [userid与openid互换](https://developer.work.weixin.qq.com/document/path/90202) — [contact/member.go:210](../contact/member.go#L210)
  - [x] [登录二次验证](https://developer.work.weixin.qq.com/document/path/90203) — [contact/member.go:240](../contact/member.go#L240)
  - [x] [邀请成员](https://developer.work.weixin.qq.com/document/path/90975) — [contact/member.go:252](../contact/member.go#L252)
  - [x] [获取加入企业二维码](https://developer.work.weixin.qq.com/document/path/91714) — [contact/member.go:269](../contact/member.go#L269)
  - [x] [手机号获取userid](https://developer.work.weixin.qq.com/document/path/95402) — [contact/member.go:284](../contact/member.go#L284)
  - [x] [邮箱获取userid](https://developer.work.weixin.qq.com/document/path/95895) — [contact/member.go:307](../contact/member.go#L307)
  - [x] [获取成员ID列表](https://developer.work.weixin.qq.com/document/path/96067) — [contact/member.go:327](../contact/member.go#L327)
- [x] 部门管理
  - [x] [创建部门](https://developer.work.weixin.qq.com/document/path/90205) — [contact/department.go:25](../contact/department.go#L25)
  - [x] [更新部门](https://developer.work.weixin.qq.com/document/path/90206) — [contact/department.go:48](../contact/department.go#L48)
  - [x] [删除部门](https://developer.work.weixin.qq.com/document/path/90207) — [contact/department.go:58](../contact/department.go#L58)
  - [x] [获取部门列表](https://developer.work.weixin.qq.com/document/path/90208) — [contact/department.go:69](../contact/department.go#L69)
  - [x] [获取子部门ID列表](https://developer.work.weixin.qq.com/document/path/95350) — [contact/department.go:84](../contact/department.go#L84)
  - [x] [获取单个部门详情](https://developer.work.weixin.qq.com/document/path/95351) — [contact/department.go:99](../contact/department.go#L99)
- [x] 标签管理
  - [x] [创建标签](https://developer.work.weixin.qq.com/document/path/90210) — [contact/tag.go:38](../contact/tag.go#L38)
  - [x] [更新标签名字](https://developer.work.weixin.qq.com/document/path/90211) — [contact/tag.go:54](../contact/tag.go#L54)
  - [x] [删除标签](https://developer.work.weixin.qq.com/document/path/90212) — [contact/tag.go:71](../contact/tag.go#L71)
  - [x] [获取标签成员](https://developer.work.weixin.qq.com/document/path/90213) — [contact/tag.go:82](../contact/tag.go#L82)
  - [x] [增加标签成员](https://developer.work.weixin.qq.com/document/path/90214) — [contact/tag.go:95](../contact/tag.go#L95)
  - [x] [删除标签成员](https://developer.work.weixin.qq.com/document/path/90215) — [contact/tag.go:102](../contact/tag.go#L102)
  - [x] [获取标签列表](https://developer.work.weixin.qq.com/document/path/90216) — [contact/tag.go:126](../contact/tag.go#L126)
- [x] 通讯录查看权限管理
  - [x] [创建规则](https://developer.work.weixin.qq.com/document/path/101536) — [contact/rule.go:33](../contact/rule.go#L33)
  - [x] [读取规则列表](https://developer.work.weixin.qq.com/document/path/101537) — [contact/rule.go:58](../contact/rule.go#L58)
  - [x] [修改规则](https://developer.work.weixin.qq.com/document/path/101539) — [contact/rule.go:69](../contact/rule.go#L69)
  - [x] [删除规则](https://developer.work.weixin.qq.com/document/path/101540) — [contact/rule.go:87](../contact/rule.go#L87)
- [x] 异步导入接口
  - [概述](https://developer.work.weixin.qq.com/document/path/90979)
  - [x] [增量更新成员](https://developer.work.weixin.qq.com/document/path/90980) — [contact/batch.go:47](../contact/batch.go#L47)
  - [x] [全量覆盖成员](https://developer.work.weixin.qq.com/document/path/90981) — [contact/batch.go:54](../contact/batch.go#L54)
  - [x] [全量覆盖部门](https://developer.work.weixin.qq.com/document/path/90982) — [contact/batch.go:61](../contact/batch.go#L61)
  - [x] [获取异步任务结果](https://developer.work.weixin.qq.com/document/path/90983) — [contact/batch.go:86](../contact/batch.go#L86)
- [x] 异步导出接口
  - [概述](https://developer.work.weixin.qq.com/document/path/94850)
  - [x] [导出成员](https://developer.work.weixin.qq.com/document/path/94849) — [contact/export.go:37](../contact/export.go#L37)
  - [x] [导出成员详情](https://developer.work.weixin.qq.com/document/path/94851) — [contact/export.go:44](../contact/export.go#L44)
  - [x] [导出部门](https://developer.work.weixin.qq.com/document/path/94852) — [contact/export.go:51](../contact/export.go#L51)
  - [x] [导出标签成员](https://developer.work.weixin.qq.com/document/path/94853) — [contact/export.go:58](../contact/export.go#L58)
  - [x] [获取导出结果](https://developer.work.weixin.qq.com/document/path/94854) — [contact/export.go:89](../contact/export.go#L89)
  - [x] [导出任务完成通知](https://developer.work.weixin.qq.com/document/path/94946) — [contact/events.go:149](../contact/events.go#L149)
- [x] 通讯录回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/90967)
  - [x] [成员变更通知](https://developer.work.weixin.qq.com/document/path/90970) — [contact/events.go:115](../contact/events.go#L115)
  - [x] [部门变更通知](https://developer.work.weixin.qq.com/document/path/90971) — [contact/events.go:115](../contact/events.go#L115)
  - [x] [标签变更通知](https://developer.work.weixin.qq.com/document/path/90972) — [contact/events.go:115](../contact/events.go#L115)
  - [x] [异步任务完成通知](https://developer.work.weixin.qq.com/document/path/90973) — [contact/events.go:149](../contact/events.go#L149)
- [通讯录同步接口调整](https://developer.work.weixin.qq.com/document/path/96079)

### 身份验证

- [x] 网页授权登录
  - [开始开发](https://developer.work.weixin.qq.com/document/path/91335)
  - [x] [构造网页授权链接](https://developer.work.weixin.qq.com/document/path/91022) — [auth/auth.go:60](../auth/auth.go#L60)
  - [x] [获取访问用户身份](https://developer.work.weixin.qq.com/document/path/91023) — [auth/auth.go:110](../auth/auth.go#L110)
  - [x] [获取访问用户敏感信息](https://developer.work.weixin.qq.com/document/path/95833) — [auth/auth.go:136](../auth/auth.go#L136)
- [x] 企业微信Web登录
  - [开始开发](https://developer.work.weixin.qq.com/document/path/98151)
  - [Web登录组件](https://developer.work.weixin.qq.com/document/path/98152)
  - [x] [获取用户登录身份](https://developer.work.weixin.qq.com/document/path/98176) — [auth/auth.go:158](../auth/auth.go#L158)
- [x] 二次验证
  - [概述](https://developer.work.weixin.qq.com/document/path/99519)
  - [x] [获取用户二次验证信息](https://developer.work.weixin.qq.com/document/path/99499) — [auth/auth.go:177](../auth/auth.go#L177)
  - [x] [登录二次验证](https://developer.work.weixin.qq.com/document/path/99521) — [auth/auth.go:205](../auth/auth.go#L205)
  - [x] [使用二次验证](https://developer.work.weixin.qq.com/document/path/99500) — [auth/auth.go:191](../auth/auth.go#L191)

### 企业互联

- [概述](https://developer.work.weixin.qq.com/document/path/93360)
- [x] [获取应用共享信息](https://developer.work.weixin.qq.com/document/path/93403) — [corpgroup/corpgroup.go:47](../corpgroup/corpgroup.go#L47)
- [x] [获取下级/下游企业的access_token](https://developer.work.weixin.qq.com/document/path/93359) — [corpgroup/corpgroup.go:77](../corpgroup/corpgroup.go#L77)
- [x] [获取下级/下游企业小程序session](https://developer.work.weixin.qq.com/document/path/93355) — [corpgroup/corpgroup.go:100](../corpgroup/corpgroup.go#L100)

### 上下游

- [概述](https://developer.work.weixin.qq.com/document/path/97213)
- [x] 基础接口
  - [x] [获取应用共享信息](https://developer.work.weixin.qq.com/document/path/95813) — [corpgroup/corpgroup.go:47](../corpgroup/corpgroup.go#L47)
  - [x] [获取下级/下游企业的access_token](https://developer.work.weixin.qq.com/document/path/95816) — [corpgroup/corpgroup.go:77](../corpgroup/corpgroup.go#L77)
  - [x] [获取下级/下游企业小程序session](https://developer.work.weixin.qq.com/document/path/95817) — [corpgroup/corpgroup.go:100](../corpgroup/corpgroup.go#L100)
  - [x] [上下游关联客户信息-已添加客户](https://developer.work.weixin.qq.com/document/path/95818) — [corpgroup/corpgroup.go:206](../corpgroup/corpgroup.go#L206)
  - [x] [上下游关联客户信息-未添加客户](https://developer.work.weixin.qq.com/document/path/97357) — [corpgroup/corpgroup.go:237](../corpgroup/corpgroup.go#L237)
- [x] 上下游通讯录管理
  - [x] [获取上下游信息](https://developer.work.weixin.qq.com/document/path/95820) — [corpgroup/corpgroup.go:129](../corpgroup/corpgroup.go#L129)
  - [x] [批量导入上下游联系人](https://developer.work.weixin.qq.com/document/path/95821) — [corpgroup/corpgroup.go:311](../corpgroup/corpgroup.go#L311)
  - [x] [获取异步任务结果](https://developer.work.weixin.qq.com/document/path/95823) — [corpgroup/corpgroup.go:355](../corpgroup/corpgroup.go#L355)
  - [x] [移除企业](https://developer.work.weixin.qq.com/document/path/95822) — [corpgroup/corpgroup.go:372](../corpgroup/corpgroup.go#L372)
  - [x] [查询成员自定义id](https://developer.work.weixin.qq.com/document/path/97441) — [corpgroup/corpgroup.go:391](../corpgroup/corpgroup.go#L391)
  - [x] [获取下级企业加入的上下游](https://developer.work.weixin.qq.com/document/path/97442) — [corpgroup/corpgroup.go:414](../corpgroup/corpgroup.go#L414)
- [x] 上下游规则
  - [x] [获取对接规则id列表](https://developer.work.weixin.qq.com/document/path/95631) — [corpgroup/corpgroup.go:447](../corpgroup/corpgroup.go#L447)
  - [x] [删除对接规则](https://developer.work.weixin.qq.com/document/path/95632) — [corpgroup/corpgroup.go:527](../corpgroup/corpgroup.go#L527)
  - [x] [获取对接规则详情](https://developer.work.weixin.qq.com/document/path/95633) — [corpgroup/corpgroup.go:466](../corpgroup/corpgroup.go#L466)
  - [x] [新增对接规则](https://developer.work.weixin.qq.com/document/path/95634) — [corpgroup/corpgroup.go:487](../corpgroup/corpgroup.go#L487)
  - [x] [更新对接规则](https://developer.work.weixin.qq.com/document/path/95635) — [corpgroup/corpgroup.go:509](../corpgroup/corpgroup.go#L509)
- [x] 回调事件
  - [概述](https://developer.work.weixin.qq.com/document/path/95794)
  - [x] [上下游变更回调](https://developer.work.weixin.qq.com/document/path/95796) — [corpgroup/corpgroup.go:562](../corpgroup/corpgroup.go#L562)
  - [x] [异步任务完成通知](https://developer.work.weixin.qq.com/document/path/95797) — [corpgroup/corpgroup.go:597](../corpgroup/corpgroup.go#L597)

### 安全管理

- [x] [文件防泄漏](https://developer.work.weixin.qq.com/document/path/98079) — [security/security.go:57](../security/security.go#L57)
- [x] [设备管理](https://developer.work.weixin.qq.com/document/path/98920) — [security/security.go:156](../security/security.go#L156)
- [x] [截屏/录屏管理](https://developer.work.weixin.qq.com/document/path/100128) — [security/security.go:98](../security/security.go#L98)
- [x] 高级功能账号管理
  - [x] [分配高级功能账号](https://developer.work.weixin.qq.com/document/path/99503) — [security/security.go:290](../security/security.go#L290)
  - [x] [取消高级功能账号](https://developer.work.weixin.qq.com/document/path/99505) — [security/security.go:313](../security/security.go#L313)
  - [x] [获取高级功能账号列表](https://developer.work.weixin.qq.com/document/path/99506) — [security/security.go:349](../security/security.go#L349)
- [x] 操作日志
  - [x] [获取成员操作记录](https://developer.work.weixin.qq.com/document/path/100178) — [security/security.go:382](../security/security.go#L382)
  - [x] [获取管理端操作日志](https://developer.work.weixin.qq.com/document/path/100179) — [security/security.go:390](../security/security.go#L390)
- [x] [获取企业微信域名IP信息](https://developer.work.weixin.qq.com/document/path/100079) — [security/security.go:428](../security/security.go#L428)
- [x] [回调通知](https://developer.work.weixin.qq.com/document/path/100080) — [security/security.go:453](../security/security.go#L453)

### 消息接收与发送

- [概述](https://developer.work.weixin.qq.com/document/path/90235)
- [x] [发送应用消息](https://developer.work.weixin.qq.com/document/path/90236) — [message/send.go:9](../message/send.go#L9)、[message/send_any.go:17](../message/send_any.go#L17)
- [x] [更新模版卡片消息](https://developer.work.weixin.qq.com/document/path/94888) — [message/recall.go:55](../message/recall.go#L55)
- [x] [撤回应用消息](https://developer.work.weixin.qq.com/document/path/94867) — [message/recall.go:17](../message/recall.go#L17)
- [x] 接收消息与事件
  - [x] [概述](https://developer.work.weixin.qq.com/document/path/90238) — [callback/callback.go:120](../callback/callback.go#L120)
  - [消息格式](https://developer.work.weixin.qq.com/document/path/90239)
  - [事件格式](https://developer.work.weixin.qq.com/document/path/90240)
  - [被动回复消息格式](https://developer.work.weixin.qq.com/document/path/90241)
- [x] 应用发送消息到群聊会话
  - [概述](https://developer.work.weixin.qq.com/document/path/90244)
  - [x] [创建群聊会话](https://developer.work.weixin.qq.com/document/path/90245) — [appchat/chat.go:20](../appchat/chat.go#L20)
  - [x] [修改群聊会话](https://developer.work.weixin.qq.com/document/path/98913) — [appchat/update.go:30](../appchat/update.go#L30)
  - [x] [获取群聊会话](https://developer.work.weixin.qq.com/document/path/98914) — [appchat/update.go:44](../appchat/update.go#L44)
  - [x] [应用推送消息](https://developer.work.weixin.qq.com/document/path/90248) — [appchat/chat.go:60](../appchat/chat.go#L60)
- 家校消息推送
  - [发送「学校通知」](https://developer.work.weixin.qq.com/document/path/91609)
- [x] 消息推送（原“群机器人”）
  - [x] [消息推送配置说明](https://developer.work.weixin.qq.com/document/path/99110) — [webhook/webhook.go:160](../webhook/webhook.go#L160)
- [x] 智能机器人
  - [概述](https://developer.work.weixin.qq.com/document/path/101039)
  - [x] [接收消息](https://developer.work.weixin.qq.com/document/path/100719) — [aibot/message.go:107](../aibot/message.go#L107)
  - [x] [接收事件](https://developer.work.weixin.qq.com/document/path/101027) — [aibot/message.go:128](../aibot/message.go#L128)
  - [x] [被动回复消息](https://developer.work.weixin.qq.com/document/path/101031) — [aibot/reply.go:65](../aibot/reply.go#L65)
  - [模板卡片类型](https://developer.work.weixin.qq.com/document/path/101032)
  - [回调和回复的加解密方案](https://developer.work.weixin.qq.com/document/path/101033)
  - [x] [主动回复消息](https://developer.work.weixin.qq.com/document/path/101138) — [aibot/aibot.go:65](../aibot/aibot.go#L65)、[aibot/reply.go:72](../aibot/reply.go#L72)
  - [x] [智能机器人长连接](https://developer.work.weixin.qq.com/document/path/101463) — [aibot/longconn/handler.go:60](../aibot/longconn/handler.go#L60)、[aibot/longconn/longconn.go:131](../aibot/longconn/longconn.go#L131)
  - [API模式机器人文档使用说明](https://developer.work.weixin.qq.com/document/path/101468)
- [x] 智能表格自动化创建的群聊
  - [x] [获取群聊列表](https://developer.work.weixin.qq.com/document/path/100989) — [wedoc/groupchat.go:43](../wedoc/groupchat.go#L43)
  - [x] [获取群聊会话](https://developer.work.weixin.qq.com/document/path/101028) — [wedoc/groupchat.go:64](../wedoc/groupchat.go#L64)
  - [x] [修改群聊会话](https://developer.work.weixin.qq.com/document/path/101029) — [wedoc/groupchat.go:85](../wedoc/groupchat.go#L85)

### 应用管理

- [概述](https://developer.work.weixin.qq.com/document/path/90226)
- [x] [获取应用](https://developer.work.weixin.qq.com/document/path/90227) — [agent/agent_info.go:39](../agent/agent_info.go#L39)
- [x] [设置应用](https://developer.work.weixin.qq.com/document/path/90228) — [agent/agent_info.go:64](../agent/agent_info.go#L64)
- [x] 自定义菜单
  - [x] [创建菜单](https://developer.work.weixin.qq.com/document/path/90231) — [agent/agent_info.go:92](../agent/agent_info.go#L92)
  - [x] [获取菜单](https://developer.work.weixin.qq.com/document/path/90232) — [agent/agent_info.go:111](../agent/agent_info.go#L111)
  - [x] [删除菜单](https://developer.work.weixin.qq.com/document/path/90233) — [agent/agent_info.go:124](../agent/agent_info.go#L124)
- [x] [设置工作台自定义展示](https://developer.work.weixin.qq.com/document/path/92535) — [agent/agent_info.go:137](../agent/agent_info.go#L137)

### 素材管理

- [概述](https://developer.work.weixin.qq.com/document/path/91054)
- [x] [上传临时素材](https://developer.work.weixin.qq.com/document/path/90253) — [media/media.go:56](../media/media.go#L56)
- [x] [上传图片](https://developer.work.weixin.qq.com/document/path/90256) — [media/media.go:73](../media/media.go#L73)
- [x] [获取临时素材](https://developer.work.weixin.qq.com/document/path/90254) — [media/media.go:90](../media/media.go#L90)
- [x] [获取高清语音素材](https://developer.work.weixin.qq.com/document/path/90255) — [media/media.go:97](../media/media.go#L97)
- [x] [异步上传临时素材](https://developer.work.weixin.qq.com/document/path/96219) — [media/media.go:115](../media/media.go#L115)

### 电子发票

- [概述](https://developer.work.weixin.qq.com/document/path/90283)
- [x] [查询电子发票](https://developer.work.weixin.qq.com/document/path/90284) — [invoice/invoice.go:71](../invoice/invoice.go#L71)
- [x] [更新发票状态](https://developer.work.weixin.qq.com/document/path/90285) — [invoice/invoice.go:108](../invoice/invoice.go#L108)
- [x] [批量更新发票状态](https://developer.work.weixin.qq.com/document/path/90286) — [invoice/invoice.go:128](../invoice/invoice.go#L128)
- [x] [批量查询电子发票](https://developer.work.weixin.qq.com/document/path/90287) — [invoice/invoice.go:87](../invoice/invoice.go#L87)

### 数据与智能专区

- [概述](https://developer.work.weixin.qq.com/document/path/99941)
- [基本概念介绍](https://developer.work.weixin.qq.com/document/path/99942)
- [接入指引](https://developer.work.weixin.qq.com/document/path/99871)
- [专区程序开发指引](https://developer.work.weixin.qq.com/document/path/100057)
- [专区程序示例](https://developer.work.weixin.qq.com/document/path/100052)
- [专区程序SDK和示例下载](https://developer.work.weixin.qq.com/document/path/100249)
- [镜像文件配置指引](https://developer.work.weixin.qq.com/document/path/99955)
- [会话展示组件](https://developer.work.weixin.qq.com/document/path/100076)
- [文件内容存档](https://developer.work.weixin.qq.com/document/path/101501)
- [x] 基础接口
  - [x] [设置公钥](https://developer.work.weixin.qq.com/document/path/99961) — [chatdata/chatdata.go:48](../chatdata/chatdata.go#L48)
  - [x] [获取会话存档授权成员列表](https://developer.work.weixin.qq.com/document/path/99962) — [chatdata/chatdata.go:63](../chatdata/chatdata.go#L63)
  - [x] [设置专区接收回调事件](https://developer.work.weixin.qq.com/document/path/99963) — [chatdata/chatdata.go:96](../chatdata/chatdata.go#L96)
  - [x] [会话组件敏感信息隐藏设置](https://developer.work.weixin.qq.com/document/path/100139) — [chatdata/chatdata.go:111](../chatdata/chatdata.go#L111)
  - [x] [设置日志打印级别](https://developer.work.weixin.qq.com/document/path/100108) — [chatdata/chatdata.go:150](../chatdata/chatdata.go#L150)
  - [x] [上传临时文件到专区](https://developer.work.weixin.qq.com/document/path/100174) — [chatdata/chatdata.go:187](../chatdata/chatdata.go#L187)
  - [x] [获取文件内容存档授权成员列表](https://developer.work.weixin.qq.com/document/path/101873) — [chatdata/chatdata.go:80](../chatdata/chatdata.go#L80)
- [x] 应用调用专区程序
  - [概述](https://developer.work.weixin.qq.com/document/path/99964)
  - [x] [应用同步调用专区程序](https://developer.work.weixin.qq.com/document/path/99965) — [chatdata/chatdata.go:206](../chatdata/chatdata.go#L206)
  - [x] [应用异步调用专区程序](https://developer.work.weixin.qq.com/document/path/99966) — [chatdata/chatdata.go:235](../chatdata/chatdata.go#L235)
- 专区程序调用sdk
  - [概述](https://developer.work.weixin.qq.com/document/path/99967)
  - [获取会话记录](https://developer.work.weixin.qq.com/document/path/99968)
  - [获取会话同意情况](https://developer.work.weixin.qq.com/document/path/99969)
  - [获取内部群信息](https://developer.work.weixin.qq.com/document/path/99970)
  - [会话名称搜索](https://developer.work.weixin.qq.com/document/path/99971)
  - [会话消息搜索](https://developer.work.weixin.qq.com/document/path/99972)
  - [员工或客户名称搜索](https://developer.work.weixin.qq.com/document/path/100243)
  - [关键词规则管理](https://developer.work.weixin.qq.com/document/path/99973)
  - [获取命中关键词规则的会话记录](https://developer.work.weixin.qq.com/document/path/99974)
  - [管理知识集](https://developer.work.weixin.qq.com/document/path/99975)
  - [通用模型](https://developer.work.weixin.qq.com/document/path/99875)
  - [话术推荐模型](https://developer.work.weixin.qq.com/document/path/99977)
  - [客户标签模型](https://developer.work.weixin.qq.com/document/path/99979)
  - [会话摘要模型](https://developer.work.weixin.qq.com/document/path/99980)
  - [情感分析模型](https://developer.work.weixin.qq.com/document/path/99983)
  - [自有模型分析](https://developer.work.weixin.qq.com/document/path/99985)
  - [会话反垃圾分析](https://developer.work.weixin.qq.com/document/path/99986)
  - [会话内容导出](https://developer.work.weixin.qq.com/document/path/99987)
  - [异步调用自有分析程序](https://developer.work.weixin.qq.com/document/path/99988)
  - [上报异步任务结果](https://developer.work.weixin.qq.com/document/path/99989)
  - [专区通知应用](https://developer.work.weixin.qq.com/document/path/99990)
- 专区程序接收事件通知
  - [概述](https://developer.work.weixin.qq.com/document/path/99992)
  - [客户同意进行聊天内容存档事件回调](https://developer.work.weixin.qq.com/document/path/99993)
  - [产生会话回调通知](https://developer.work.weixin.qq.com/document/path/99994)
  - [命中关键词规则通知](https://developer.work.weixin.qq.com/document/path/99995)
  - [知识集管理回调](https://developer.work.weixin.qq.com/document/path/99996)
  - [会话内容导出完成通知](https://developer.work.weixin.qq.com/document/path/99997)
- [x] 应用接收专区通知
  - [x] [应用接收专区通知](https://developer.work.weixin.qq.com/document/path/99998) — [chatdata/chatdata.go:351](../chatdata/chatdata.go#L351)
- [x] 专区调试模式
  - [调试说明](https://developer.work.weixin.qq.com/document/path/100086)
  - [x] [开启专区调试模式](https://developer.work.weixin.qq.com/document/path/100087) — [chatdata/chatdata.go:299](../chatdata/chatdata.go#L299)
  - [x] [关闭专区调试模式](https://developer.work.weixin.qq.com/document/path/100088) — [chatdata/chatdata.go:316](../chatdata/chatdata.go#L316)
  - [x] [获取专区调试模式状态](https://developer.work.weixin.qq.com/document/path/100113) — [chatdata/chatdata.go:280](../chatdata/chatdata.go#L280)
- [常见问题](https://developer.work.weixin.qq.com/document/path/100142)

## 连接微信

### 客户联系

- [概述](https://developer.work.weixin.qq.com/document/path/92109)
- [x] [成员对外信息](https://developer.work.weixin.qq.com/document/path/92230) — [external/customer.go:151](../external/customer.go#L151)
- [x] 企业服务人员管理
  - [x] [获取配置了客户联系功能的成员列表](https://developer.work.weixin.qq.com/document/path/92571) — [external/customer.go:71](../external/customer.go#L71)
- [x] 客户管理
  - [x] [获取客户列表](https://developer.work.weixin.qq.com/document/path/92113) — [external/customer.go:82](../external/customer.go#L82)
  - [x] [获取客户详情](https://developer.work.weixin.qq.com/document/path/92114) — [external/customer.go:100](../external/customer.go#L100)
  - [x] [批量获取客户详情](https://developer.work.weixin.qq.com/document/path/92994) — [external/customer.go:123](../external/customer.go#L123)
  - [x] [修改客户备注信息](https://developer.work.weixin.qq.com/document/path/92115) — [external/customer.go:140](../external/customer.go#L140)
  - [x] [客户联系规则组管理](https://developer.work.weixin.qq.com/document/path/94883) — [external/strategy.go:72](../external/strategy.go#L72)
- [x] 客户标签管理
  - [x] [管理企业标签](https://developer.work.weixin.qq.com/document/path/92117) — [external/tag.go:40](../external/tag.go#L40)
  - [x] [管理企业规则组下的客户标签](https://developer.work.weixin.qq.com/document/path/94882) — [external/moment_strategy.go:115](../external/moment_strategy.go#L115)
  - [x] [编辑客户企业标签](https://developer.work.weixin.qq.com/document/path/92118) — [external/tag.go:101](../external/tag.go#L101)
- [x] 在职继承
  - [x] [分配在职成员的客户](https://developer.work.weixin.qq.com/document/path/92125) — [external/transfer.go:41](../external/transfer.go#L41)
  - [x] [查询客户接替状态](https://developer.work.weixin.qq.com/document/path/94088) — [external/transfer.go:58](../external/transfer.go#L58)
  - [x] [分配在职成员的客户群](https://developer.work.weixin.qq.com/document/path/95703) — [external/transfer.go:92](../external/transfer.go#L92)
- [x] 离职继承
  - [x] [获取待分配的离职成员列表](https://developer.work.weixin.qq.com/document/path/92124) — [external/transfer.go:132](../external/transfer.go#L132)
  - [x] [分配离职成员的客户](https://developer.work.weixin.qq.com/document/path/94081) — [external/transfer.go:169](../external/transfer.go#L169)
  - [x] [查询客户接替状态](https://developer.work.weixin.qq.com/document/path/94082) — [external/transfer.go:150](../external/transfer.go#L150)
  - [x] [分配离职成员的客户群](https://developer.work.weixin.qq.com/document/path/92127) — [external/transfer.go:84](../external/transfer.go#L84)
- [x] 客户群管理
  - [x] [获取客户群列表](https://developer.work.weixin.qq.com/document/path/92120) — [external/groupchat.go:64](../external/groupchat.go#L64)
  - [x] [获取客户群详情](https://developer.work.weixin.qq.com/document/path/92122) — [external/groupchat.go:80](../external/groupchat.go#L80)
  - [x] [客户群opengid转换](https://developer.work.weixin.qq.com/document/path/94822) — [external/groupchat.go:99](../external/groupchat.go#L99)
- [x] 联系我与客户入群方式
  - [x] [客户联系「联系我」管理](https://developer.work.weixin.qq.com/document/path/92228) — [external/groupchat.go:199](../external/groupchat.go#L199)
  - [x] [客户群「加入群聊」管理](https://developer.work.weixin.qq.com/document/path/92229) — [external/groupchat.go:128](../external/groupchat.go#L128)
- [x] 客户朋友圈
  - [概述](https://developer.work.weixin.qq.com/document/path/93506)
  - [x] [企业发表内容到客户的朋友圈](https://developer.work.weixin.qq.com/document/path/95094) — [external/moment.go:68](../external/moment.go#L68)
  - [x] [停止发表企业朋友圈](https://developer.work.weixin.qq.com/document/path/97612) — [external/moment.go:98](../external/moment.go#L98)
  - [x] [获取客户朋友圈全部的发表记录](https://developer.work.weixin.qq.com/document/path/93333) — [external/moment_record.go:55](../external/moment_record.go#L55)
  - [x] [客户朋友圈规则组管理](https://developer.work.weixin.qq.com/document/path/94890) — [external/moment_strategy.go:29](../external/moment_strategy.go#L29)
- [x] 获客助手
  - [x] [获客链接管理](https://developer.work.weixin.qq.com/document/path/97297) — [external/acquisition.go:51](../external/acquisition.go#L51)
  - [x] [获取由获客链接添加的客户信息](https://developer.work.weixin.qq.com/document/path/97298) — [external/acquisition.go:135](../external/acquisition.go#L135)
  - [x] [获客助手额度管理与使用统计](https://developer.work.weixin.qq.com/document/path/97375) — [external/acquisition.go:166](../external/acquisition.go#L166)
  - [提升广告有效率](https://developer.work.weixin.qq.com/document/path/99596)
  - [x] [获取成员多次收消息详情](https://developer.work.weixin.qq.com/document/path/100130) — [external/acquisition.go:215](../external/acquisition.go#L215)
- [x] 消息推送
  - [x] [创建企业群发](https://developer.work.weixin.qq.com/document/path/92135) — [external/message.go:83](../external/message.go#L83)
  - [x] [提醒成员群发](https://developer.work.weixin.qq.com/document/path/97610) — [external/message.go:98](../external/message.go#L98)
  - [x] [停止企业群发](https://developer.work.weixin.qq.com/document/path/97611) — [external/message.go:109](../external/message.go#L109)
  - [x] [获取企业的全部群发记录](https://developer.work.weixin.qq.com/document/path/93338) — [external/message.go:147](../external/message.go#L147)
  - [x] [发送新客户欢迎语](https://developer.work.weixin.qq.com/document/path/92137) — [external/message.go:214](../external/message.go#L214)
  - [x] [入群欢迎语素材管理](https://developer.work.weixin.qq.com/document/path/92366) — [external/welcome_template.go:25](../external/welcome_template.go#L25)
- [x] 统计管理
  - [x] [获取「联系客户统计」数据](https://developer.work.weixin.qq.com/document/path/92132) — [external/statistic.go:25](../external/statistic.go#L25)
  - [x] [获取「群聊数据统计」数据](https://developer.work.weixin.qq.com/document/path/92133) — [external/statistic.go:82](../external/statistic.go#L82)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/92129)
  - [事件格式](https://developer.work.weixin.qq.com/document/path/92130)
  - [x] [获客助手事件通知](https://developer.work.weixin.qq.com/document/path/97299) — [external/acquisition_event.go:45](../external/acquisition_event.go#L45)
- [x] [管理商品图册](https://developer.work.weixin.qq.com/document/path/95096) — [external/resource.go:34](../external/resource.go#L34)
- [x] [管理聊天敏感词](https://developer.work.weixin.qq.com/document/path/95097) — [external/resource.go:122](../external/resource.go#L122)
- [x] [上传附件资源](https://developer.work.weixin.qq.com/document/path/95098) — [external/resource.go:221](../external/resource.go#L221)
- [x] [获取已服务的外部联系人](https://developer.work.weixin.qq.com/document/path/99434) — [external/resource.go:260](../external/resource.go#L260)

### 微信客服

- [概述](https://developer.work.weixin.qq.com/document/path/94638)
- [x] 客服账号管理
  - [x] [添加客服账号](https://developer.work.weixin.qq.com/document/path/94662) — [kf/account.go:42](../kf/account.go#L42)
  - [x] [删除客服账号](https://developer.work.weixin.qq.com/document/path/94663) — [kf/account.go:61](../kf/account.go#L61)
  - [x] [修改客服账号](https://developer.work.weixin.qq.com/document/path/94664) — [kf/account.go:69](../kf/account.go#L69)
  - [x] [获取客服账号列表](https://developer.work.weixin.qq.com/document/path/94661) — [kf/account.go:23](../kf/account.go#L23)
  - [x] [获取客服账号链接](https://developer.work.weixin.qq.com/document/path/94665) — [kf/account.go:92](../kf/account.go#L92)
- [x] 接待人员管理
  - [x] [添加接待人员](https://developer.work.weixin.qq.com/document/path/94646) — [kf/kf.go:269](../kf/kf.go#L269)
  - [x] [删除接待人员](https://developer.work.weixin.qq.com/document/path/94647) — [kf/kf.go:276](../kf/kf.go#L276)
  - [x] [获取接待人员列表](https://developer.work.weixin.qq.com/document/path/94645) — [kf/kf.go:242](../kf/kf.go#L242)
- [x] 会话分配与消息收发
  - [x] [分配客服会话](https://developer.work.weixin.qq.com/document/path/94669) — [kf/kf.go:41](../kf/kf.go#L41)
  - [x] [接收消息和事件](https://developer.work.weixin.qq.com/document/path/94670) — [kf/kf.go:171](../kf/kf.go#L171)
  - [x] [发送消息](https://developer.work.weixin.qq.com/document/path/94677) — [kf/kf.go:211](../kf/kf.go#L211)
  - [x] [发送欢迎语等事件响应消息](https://developer.work.weixin.qq.com/document/path/95122) — [kf/event.go:69](../kf/event.go#L69)
- [x] [「升级服务」配置](https://developer.work.weixin.qq.com/document/path/94674) — [kf/upgrade.go:60](../kf/upgrade.go#L60)
- [x] 其他基础信息获取
  - [x] [获取客户基础信息](https://developer.work.weixin.qq.com/document/path/95159) — [kf/customer.go:35](../kf/customer.go#L35)
- [x] 统计管理
  - [x] [获取「客户数据统计」企业汇总数据](https://developer.work.weixin.qq.com/document/path/95489) — [kf/statistic.go:57](../kf/statistic.go#L57)
  - [x] [获取「客户数据统计」接待人员明细数据](https://developer.work.weixin.qq.com/document/path/95490) — [kf/statistic.go:76](../kf/statistic.go#L76)
- [x] 机器人管理
  - [x] [知识库分组管理](https://developer.work.weixin.qq.com/document/path/95971) — [kf/knowledge.go:80](../kf/knowledge.go#L80)
  - [x] [知识库问答管理](https://developer.work.weixin.qq.com/document/path/95972) — [kf/knowledge.go:138](../kf/knowledge.go#L138)
- [x] [回调通知](https://developer.work.weixin.qq.com/document/path/97712) — [kf/events.go:30](../kf/events.go#L30)

### 企业支付

- [概述](https://developer.work.weixin.qq.com/document/path/90273)
- [x] 企业红包
  - [x] [发放企业红包](https://developer.work.weixin.qq.com/document/path/90275) — [payment/payment.go:238](../payment/payment.go#L238)
  - [x] [查询红包记录](https://developer.work.weixin.qq.com/document/path/90276) — [payment/payment.go:273](../payment/payment.go#L273)
- [x] 向员工付款
  - [x] [向员工付款](https://developer.work.weixin.qq.com/document/path/90278) — [payment/payment.go:301](../payment/payment.go#L301)
  - [x] [查询付款记录](https://developer.work.weixin.qq.com/document/path/90279) — [payment/payment.go:335](../payment/payment.go#L335)
- [向员工收款](https://developer.work.weixin.qq.com/document/path/90280)
- [x] 对外收款
  - [概述](https://developer.work.weixin.qq.com/document/path/93665)
  - [x] [收款商户号管理](https://developer.work.weixin.qq.com/document/path/93666) — [externalpay/externalpay.go:39](../externalpay/externalpay.go#L39)
  - [x] [获取对外收款记录](https://developer.work.weixin.qq.com/document/path/93667) — [externalpay/externalpay.go:94](../externalpay/externalpay.go#L94)
  - [x] [获取收款项目的商户单号](https://developer.work.weixin.qq.com/document/path/95944) — [externalpay/externalpay.go:116](../externalpay/externalpay.go#L116)
  - [x] [获取资金流水](https://developer.work.weixin.qq.com/document/path/98100) — [externalpay/externalpay.go:162](../externalpay/externalpay.go#L162)
- [签名算法](https://developer.work.weixin.qq.com/document/path/90281)
- [x] 创建对外收款账户
  - [x] [提交创建对外收款账户的申请单](https://developer.work.weixin.qq.com/document/path/98973) — [externalpay/externalpay.go:192](../externalpay/externalpay.go#L192)
  - [x] [查询申请单状态](https://developer.work.weixin.qq.com/document/path/98974) — [externalpay/externalpay.go:226](../externalpay/externalpay.go#L226)
  - [x] [提交图片](https://developer.work.weixin.qq.com/document/path/98972) — [externalpay/externalpay.go:246](../externalpay/externalpay.go#L246)

### 小程序接入对外收款

- [概述](https://developer.work.weixin.qq.com/document/path/98723)
- [x] 普通支付
  - [x] [小程序下单](https://developer.work.weixin.qq.com/document/path/97322) — [miniapppay/miniapppay.go:127](../miniapppay/miniapppay.go#L127)
  - [x] [查询订单](https://developer.work.weixin.qq.com/document/path/97323) — [miniapppay/miniapppay.go:149](../miniapppay/miniapppay.go#L149)
  - [x] [关闭订单](https://developer.work.weixin.qq.com/document/path/97324) — [miniapppay/miniapppay.go:165](../miniapppay/miniapppay.go#L165)
  - [x] [获取支付签名](https://developer.work.weixin.qq.com/document/path/98130) — [miniapppay/miniapppay.go:179](../miniapppay/miniapppay.go#L179)
- [x] 退款
  - [x] [申请退款](https://developer.work.weixin.qq.com/document/path/97333) — [miniapppay/miniapppay.go:203](../miniapppay/miniapppay.go#L203)
  - [x] [查询退款](https://developer.work.weixin.qq.com/document/path/97352) — [miniapppay/miniapppay.go:221](../miniapppay/miniapppay.go#L221)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/97478)
  - [x] [支付通知](https://developer.work.weixin.qq.com/document/path/97335) — [miniapppay/miniapppay.go:305](../miniapppay/miniapppay.go#L305)
  - [x] [退款通知](https://developer.work.weixin.qq.com/document/path/97338) — [miniapppay/miniapppay.go:320](../miniapppay/miniapppay.go#L320)
- [x] 账单
  - [x] [交易账单申请](https://developer.work.weixin.qq.com/document/path/98115) — [miniapppay/miniapppay.go:238](../miniapppay/miniapppay.go#L238)

### 会话内容存档

- [概述](https://developer.work.weixin.qq.com/document/path/91360)
- [使用前帮助](https://developer.work.weixin.qq.com/document/path/91361)
- [x] [获取会话内容](https://developer.work.weixin.qq.com/document/path/91774) — [msgaudit/msgaudit.go:51](../msgaudit/msgaudit.go#L51)
- [开发案例演示](https://developer.work.weixin.qq.com/document/path/91551)
- [常见问题解答](https://developer.work.weixin.qq.com/document/path/91552)
- [x] [获取会话内容存档开启成员列表](https://developer.work.weixin.qq.com/document/path/91614) — [msgaudit/msgaudit.go:139](../msgaudit/msgaudit.go#L139)
- [x] [获取会话同意情况](https://developer.work.weixin.qq.com/document/path/91782) — [msgaudit/msgaudit.go:163](../msgaudit/msgaudit.go#L163)
- [x] [客户同意进行聊天内容存档事件回调](https://developer.work.weixin.qq.com/document/path/92005) — [msgaudit/msgaudit.go:263](../msgaudit/msgaudit.go#L263)
- [x] [获取会话内容存档内部群信息](https://developer.work.weixin.qq.com/document/path/92951) — [msgaudit/msgaudit.go:223](../msgaudit/msgaudit.go#L223)
- [x] [产生会话回调事件](https://developer.work.weixin.qq.com/document/path/95039) — [msgaudit/msgaudit.go:277](../msgaudit/msgaudit.go#L277)

### 家校沟通

- [概述](https://developer.work.weixin.qq.com/document/path/91638)
- 基础接口
  - [获取「学校通知」二维码](https://developer.work.weixin.qq.com/document/path/92320)
  - [管理「学校通知」的关注模式](https://developer.work.weixin.qq.com/document/path/92318)
  - [发送「学校通知」](https://developer.work.weixin.qq.com/document/path/92321)
  - [管理「班级群创建方式」](https://developer.work.weixin.qq.com/document/path/92430)
  - [外部联系人openid转换](https://developer.work.weixin.qq.com/document/path/92323)
  - [获取可使用的家长范围](https://developer.work.weixin.qq.com/document/path/94895)
- 网页授权登录
  - [开始开发](https://developer.work.weixin.qq.com/document/path/91856)
  - [构造网页授权链接](https://developer.work.weixin.qq.com/document/path/91857)
  - [获取访问用户身份](https://developer.work.weixin.qq.com/document/path/91707)
  - [获取家校访问用户身份](https://developer.work.weixin.qq.com/document/path/95791)
- 学生与家长管理
  - [创建学生](https://developer.work.weixin.qq.com/document/path/92325)
  - [删除学生](https://developer.work.weixin.qq.com/document/path/92326)
  - [更新学生](https://developer.work.weixin.qq.com/document/path/92327)
  - [批量创建学生](https://developer.work.weixin.qq.com/document/path/92328)
  - [批量删除学生](https://developer.work.weixin.qq.com/document/path/92329)
  - [批量更新学生](https://developer.work.weixin.qq.com/document/path/92330)
  - [创建家长](https://developer.work.weixin.qq.com/document/path/92331)
  - [删除家长](https://developer.work.weixin.qq.com/document/path/92332)
  - [更新家长](https://developer.work.weixin.qq.com/document/path/92333)
  - [批量创建家长](https://developer.work.weixin.qq.com/document/path/92334)
  - [批量删除家长](https://developer.work.weixin.qq.com/document/path/92335)
  - [批量更新家长](https://developer.work.weixin.qq.com/document/path/92336)
  - [读取学生或家长](https://developer.work.weixin.qq.com/document/path/92337)
  - [获取部门学生详情](https://developer.work.weixin.qq.com/document/path/92338)
  - [设置家校通讯录自动同步模式](https://developer.work.weixin.qq.com/document/path/92345)
  - [获取部门家长详情](https://developer.work.weixin.qq.com/document/path/92446)
- 部门管理
  - [创建部门](https://developer.work.weixin.qq.com/document/path/92340)
  - [更新部门](https://developer.work.weixin.qq.com/document/path/92341)
  - [删除部门](https://developer.work.weixin.qq.com/document/path/92342)
  - [获取部门列表](https://developer.work.weixin.qq.com/document/path/92343)
  - [标准年级对照表](https://developer.work.weixin.qq.com/document/path/92344)
  - [修改自动升年级的配置](https://developer.work.weixin.qq.com/document/path/92949)
- 家校通讯录变更回调
  - [成员变更事件](https://developer.work.weixin.qq.com/document/path/92032)
  - [部门变更事件](https://developer.work.weixin.qq.com/document/path/92052)

### 家校应用

- 健康上报
  - [获取健康上报使用统计](https://developer.work.weixin.qq.com/document/path/93676)
  - [获取健康上报任务ID列表](https://developer.work.weixin.qq.com/document/path/93677)
  - [获取健康上报任务详情](https://developer.work.weixin.qq.com/document/path/93678)
  - [获取用户填写答案](https://developer.work.weixin.qq.com/document/path/93679)
- 上课直播
  - [获取老师直播ID列表](https://developer.work.weixin.qq.com/document/path/93739)
  - [获取直播详情](https://developer.work.weixin.qq.com/document/path/93740)
  - [获取观看直播统计](https://developer.work.weixin.qq.com/document/path/93741)
  - [获取未观看直播统计](https://developer.work.weixin.qq.com/document/path/93742)
  - [删除直播回放](https://developer.work.weixin.qq.com/document/path/93743)
  - [获取观看直播统计V2](https://developer.work.weixin.qq.com/document/path/95793)
  - [获取未观看直播统计V2](https://developer.work.weixin.qq.com/document/path/95795)
- 班级收款
  - [获取学生付款结果](https://developer.work.weixin.qq.com/document/path/94470)
  - [获取订单详情](https://developer.work.weixin.qq.com/document/path/94471)

### 政民沟通

- 配置网格结构
  - [概述](https://developer.work.weixin.qq.com/document/path/94557)
  - [添加网格](https://developer.work.weixin.qq.com/document/path/94478)
  - [编辑网格](https://developer.work.weixin.qq.com/document/path/94479)
  - [删除网格](https://developer.work.weixin.qq.com/document/path/94480)
  - [获取网格列表](https://developer.work.weixin.qq.com/document/path/94481)
  - [获取用户负责及参与的网格列表](https://developer.work.weixin.qq.com/document/path/94482)
- 配置事件类别
  - [添加事件类别](https://developer.work.weixin.qq.com/document/path/94536)
  - [修改事件类别](https://developer.work.weixin.qq.com/document/path/94537)
  - [删除事件类别](https://developer.work.weixin.qq.com/document/path/94538)
  - [获取事件类别列表](https://developer.work.weixin.qq.com/document/path/94540)
- 巡查上报
  - [概述](https://developer.work.weixin.qq.com/document/path/93520)
  - [获取配置的网格及网格负责人](https://developer.work.weixin.qq.com/document/path/93531)
  - [获取单位巡查上报数据统计](https://developer.work.weixin.qq.com/document/path/93532)
  - [获取个人巡查上报数据统计](https://developer.work.weixin.qq.com/document/path/93533)
  - [获取上报事件分类统计](https://developer.work.weixin.qq.com/document/path/93534)
  - [获取巡查上报事件列表](https://developer.work.weixin.qq.com/document/path/93536)
  - [获取巡查上报的事件详情信息](https://developer.work.weixin.qq.com/document/path/93535)
- 居民上报
  - [概述](https://developer.work.weixin.qq.com/document/path/93513)
  - [获取配置的网格及网格负责人](https://developer.work.weixin.qq.com/document/path/93514)
  - [获取单位居民上报数据统计](https://developer.work.weixin.qq.com/document/path/93515)
  - [获取个人居民上报数据统计](https://developer.work.weixin.qq.com/document/path/93516)
  - [获取上报事件分类统计](https://developer.work.weixin.qq.com/document/path/93517)
  - [获取居民上报事件列表](https://developer.work.weixin.qq.com/document/path/93518)
  - [获取居民上报的事件详情信息](https://developer.work.weixin.qq.com/document/path/93519)

## 办公

### 邮件

- [概述](https://developer.work.weixin.qq.com/document/path/95486)
- [x] 发送邮件
  - [x] [发送普通邮件](https://developer.work.weixin.qq.com/document/path/97445) — [mail/send.go:85](../mail/send.go#L85)
  - [x] [发送日程邮件](https://developer.work.weixin.qq.com/document/path/97854) — [mail/send.go:85](../mail/send.go#L85)
  - [x] [发送会议邮件](https://developer.work.weixin.qq.com/document/path/97855) — [mail/send.go:85](../mail/send.go#L85)
- [x] 获取接收的邮件
  - [x] [获取收件箱邮件列表](https://developer.work.weixin.qq.com/document/path/97369) — [mail/inbox.go:27](../mail/inbox.go#L27)
  - [x] [获取邮件内容](https://developer.work.weixin.qq.com/document/path/97979) — [mail/inbox.go:42](../mail/inbox.go#L42)
- [x] 管理应用邮箱账号
  - [x] [更新应用邮箱账号](https://developer.work.weixin.qq.com/document/path/97373) — [mail/account.go:52](../mail/account.go#L52)
  - [x] [查询应用邮箱账号](https://developer.work.weixin.qq.com/document/path/97991) — [mail/account.go:43](../mail/account.go#L43)
- [x] 管理邮件群组
  - [x] [创建邮件群组](https://developer.work.weixin.qq.com/document/path/95510) — [mail/group.go:32](../mail/group.go#L32)
  - [x] [更新邮件群组](https://developer.work.weixin.qq.com/document/path/97995) — [mail/group.go:42](../mail/group.go#L42)
  - [x] [删除邮件群组](https://developer.work.weixin.qq.com/document/path/97996) — [mail/group.go:52](../mail/group.go#L52)
  - [x] [获取邮件群组详情](https://developer.work.weixin.qq.com/document/path/97997) — [mail/group.go:63](../mail/group.go#L63)
  - [x] [模糊搜索邮件群组](https://developer.work.weixin.qq.com/document/path/97998) — [mail/group.go:77](../mail/group.go#L77)
- [x] 管理公共邮箱
  - [x] [创建公共邮箱](https://developer.work.weixin.qq.com/document/path/95511) — [mail/publicmail.go:48](../mail/publicmail.go#L48)
  - [x] [更新公共邮箱](https://developer.work.weixin.qq.com/document/path/98000) — [mail/publicmail.go:58](../mail/publicmail.go#L58)
  - [x] [删除公共邮箱](https://developer.work.weixin.qq.com/document/path/98001) — [mail/publicmail.go:68](../mail/publicmail.go#L68)
  - [x] [获取公共邮箱详情](https://developer.work.weixin.qq.com/document/path/98002) — [mail/publicmail.go:78](../mail/publicmail.go#L78)
  - [x] [模糊搜索公共邮箱](https://developer.work.weixin.qq.com/document/path/98003) — [mail/publicmail.go:93](../mail/publicmail.go#L93)
  - [x] [回调通知](https://developer.work.weixin.qq.com/document/path/100180) — [mail/events.go:39](../mail/events.go#L39)
  - [x] [获取客户端专用密码列表](https://developer.work.weixin.qq.com/document/path/100183) — [mail/publicmail.go:110](../mail/publicmail.go#L110)
  - [x] [删除客户端专用密码](https://developer.work.weixin.qq.com/document/path/100184) — [mail/publicmail.go:124](../mail/publicmail.go#L124)
- [x] 高级功能账号管理
  - [x] [分配高级功能账号](https://developer.work.weixin.qq.com/document/path/99316) — [mail/vip.go:28](../mail/vip.go#L28)
  - [x] [取消高级功能账号](https://developer.work.weixin.qq.com/document/path/99317) — [mail/vip.go:40](../mail/vip.go#L40)
  - [x] [获取高级功能账号列表](https://developer.work.weixin.qq.com/document/path/99318) — [mail/vip.go:52](../mail/vip.go#L52)
- [x] [禁用/启用邮箱账号](https://developer.work.weixin.qq.com/document/path/95512) — [mail/account.go:64](../mail/account.go#L64)
- [x] 其他邮件客户端登录设置
  - [x] [获取用户功能属性](https://developer.work.weixin.qq.com/document/path/95513) — [mail/account.go:82](../mail/account.go#L82)
  - [x] [更改用户功能属性](https://developer.work.weixin.qq.com/document/path/98008) — [mail/account.go:104](../mail/account.go#L104)
- [x] [获取邮件未读数](https://developer.work.weixin.qq.com/document/path/95514) — [mail/inbox.go:57](../mail/inbox.go#L57)
- [x] [回调通知](https://developer.work.weixin.qq.com/document/path/97495) — [mail/events.go:39](../mail/events.go#L39)

### 文档

- [概述](https://developer.work.weixin.qq.com/document/path/97392)
- [x] 管理文档
  - [x] [新建文档](https://developer.work.weixin.qq.com/document/path/97460) — [wedoc/doc.go:54](../wedoc/doc.go#L54)
  - [x] [重命名文档](https://developer.work.weixin.qq.com/document/path/97736) — [wedoc/doc.go:103](../wedoc/doc.go#L103)
  - [x] [删除文档](https://developer.work.weixin.qq.com/document/path/97735) — [wedoc/doc.go:313](../wedoc/doc.go#L313)
  - [x] [获取文档基础信息](https://developer.work.weixin.qq.com/document/path/97734) — [wedoc/doc.go:131](../wedoc/doc.go#L131)
  - [x] [分享文档](https://developer.work.weixin.qq.com/document/path/97733) — [wedoc/doc.go:151](../wedoc/doc.go#L151)
- [x] 管理文档内容
  - [x] [编辑文档内容](https://developer.work.weixin.qq.com/document/path/97626) — [wedoc/document.go:25](../wedoc/document.go#L25)
  - [x] [获取文档数据](https://developer.work.weixin.qq.com/document/path/101161) — [wedoc/document.go:44](../wedoc/document.go#L44)
- [x] 管理表格内容
  - [x] [编辑表格内容](https://developer.work.weixin.qq.com/document/path/101168) — [wedoc/spreadsheet.go:20](../wedoc/spreadsheet.go#L20)
  - [x] [获取表格行列信息](https://developer.work.weixin.qq.com/document/path/97711) — [wedoc/spreadsheet.go:44](../wedoc/spreadsheet.go#L44)
  - [x] [获取表格数据](https://developer.work.weixin.qq.com/document/path/97661) — [wedoc/spreadsheet.go:59](../wedoc/spreadsheet.go#L59)
- [x] 管理智能表格内容
  - [x] [添加子表](https://developer.work.weixin.qq.com/document/path/99896) — [wedoc/sheet.go:37](../wedoc/sheet.go#L37)
  - [x] [删除子表](https://developer.work.weixin.qq.com/document/path/99899) — [wedoc/sheet.go:96](../wedoc/sheet.go#L96)
  - [x] [更新子表](https://developer.work.weixin.qq.com/document/path/99898) — [wedoc/sheet.go:73](../wedoc/sheet.go#L73)
  - [x] [查询子表](https://developer.work.weixin.qq.com/document/path/101154) — [wedoc/sheet.go:116](../wedoc/sheet.go#L116)
  - [x] [添加视图](https://developer.work.weixin.qq.com/document/path/99900) — [wedoc/view.go:41](../wedoc/view.go#L41)
  - [x] [删除视图](https://developer.work.weixin.qq.com/document/path/99901) — [wedoc/view.go:123](../wedoc/view.go#L123)
  - [x] [更新视图](https://developer.work.weixin.qq.com/document/path/99902) — [wedoc/view.go:85](../wedoc/view.go#L85)
  - [x] [查询视图](https://developer.work.weixin.qq.com/document/path/101155) — [wedoc/view.go:154](../wedoc/view.go#L154)
  - [x] [添加字段](https://developer.work.weixin.qq.com/document/path/99904) — [wedoc/field.go:146](../wedoc/field.go#L146)
  - [x] [删除字段](https://developer.work.weixin.qq.com/document/path/99905) — [wedoc/field.go:235](../wedoc/field.go#L235)
  - [x] [更新字段](https://developer.work.weixin.qq.com/document/path/99906) — [wedoc/field.go:198](../wedoc/field.go#L198)
  - [x] [查询字段](https://developer.work.weixin.qq.com/document/path/101157) — [wedoc/field.go:266](../wedoc/field.go#L266)
  - [x] [添加记录](https://developer.work.weixin.qq.com/document/path/99907) — [wedoc/record.go:70](../wedoc/record.go#L70)
  - [x] [删除记录](https://developer.work.weixin.qq.com/document/path/99908) — [wedoc/record.go:128](../wedoc/record.go#L128)
  - [x] [更新记录](https://developer.work.weixin.qq.com/document/path/99909) — [wedoc/record.go:97](../wedoc/record.go#L97)
  - [x] [查询记录](https://developer.work.weixin.qq.com/document/path/101158) — [wedoc/record.go:184](../wedoc/record.go#L184)
  - [x] [添加编组](https://developer.work.weixin.qq.com/document/path/101100) — [wedoc/field_group.go:33](../wedoc/field_group.go#L33)
  - [x] [删除编组](https://developer.work.weixin.qq.com/document/path/101102) — [wedoc/field_group.go:115](../wedoc/field_group.go#L115)
  - [x] [更新编组](https://developer.work.weixin.qq.com/document/path/101101) — [wedoc/field_group.go:69](../wedoc/field_group.go#L69)
  - [x] [获取编组](https://developer.work.weixin.qq.com/document/path/101103) — [wedoc/field_group.go:147](../wedoc/field_group.go#L147)
- [x] 管理智能文档内容
  - [x] [添加页面](https://developer.work.weixin.qq.com/document/path/101620) — [wedoc/smartdoc.go:99](../wedoc/smartdoc.go#L99)
  - [x] [更新页面](https://developer.work.weixin.qq.com/document/path/101621) — [wedoc/smartdoc.go:118](../wedoc/smartdoc.go#L118)
  - [x] [删除页面](https://developer.work.weixin.qq.com/document/path/101622) — [wedoc/smartdoc.go:140](../wedoc/smartdoc.go#L140)
  - [x] [获取页面结构](https://developer.work.weixin.qq.com/document/path/101619) — [wedoc/smartdoc.go:84](../wedoc/smartdoc.go#L84)
  - [x] [添加内容块](https://developer.work.weixin.qq.com/document/path/101623) — [wedoc/smartdoc.go:155](../wedoc/smartdoc.go#L155)
  - [x] [更新内容块](https://developer.work.weixin.qq.com/document/path/101624) — [wedoc/smartdoc.go:177](../wedoc/smartdoc.go#L177)
  - [x] [删除内容块](https://developer.work.weixin.qq.com/document/path/101625) — [wedoc/smartdoc.go:199](../wedoc/smartdoc.go#L199)
  - [x] [获取内容块列表](https://developer.work.weixin.qq.com/document/path/101626) — [wedoc/smartdoc.go:218](../wedoc/smartdoc.go#L218)
  - [x] [导出内容块](https://developer.work.weixin.qq.com/document/path/101627) — [wedoc/smartdoc.go:241](../wedoc/smartdoc.go#L241)
  - [x] [获取数据源](https://developer.work.weixin.qq.com/document/path/101628) — [wedoc/smartdoc.go:275](../wedoc/smartdoc.go#L275)
  - [x] [添加数据表](https://developer.work.weixin.qq.com/document/path/101629) — [wedoc/smartdoc.go:290](../wedoc/smartdoc.go#L290)
  - [x] [删除数据表](https://developer.work.weixin.qq.com/document/path/101630) — [wedoc/smartdoc.go:309](../wedoc/smartdoc.go#L309)
  - [x] [更新数据表](https://developer.work.weixin.qq.com/document/path/101631) — [wedoc/smartdoc.go:324](../wedoc/smartdoc.go#L324)
  - [x] [发布智能文档](https://developer.work.weixin.qq.com/document/path/101616) — [wedoc/smartdoc.go:347](../wedoc/smartdoc.go#L347)
  - [x] [取消发布智能文档](https://developer.work.weixin.qq.com/document/path/101617) — [wedoc/smartdoc.go:365](../wedoc/smartdoc.go#L365)
  - [x] [修改发布页可查看范围](https://developer.work.weixin.qq.com/document/path/101618) — [wedoc/smartdoc.go:376](../wedoc/smartdoc.go#L376)
- [x] 设置文档权限
  - [x] [获取文档权限信息](https://developer.work.weixin.qq.com/document/path/97461) — [wedoc/doc.go:224](../wedoc/doc.go#L224)
  - [x] [修改文档加入规则](https://developer.work.weixin.qq.com/document/path/97778) — [wedoc/doc.go:257](../wedoc/doc.go#L257)
  - [x] [修改文档成员与权限](https://developer.work.weixin.qq.com/document/path/97781) — [wedoc/doc.go:350](../wedoc/doc.go#L350)
  - [x] [修改文档安全设置](https://developer.work.weixin.qq.com/document/path/97782) — [wedoc/security.go:24](../wedoc/security.go#L24)
  - [x] [管理智能表格内容权限](https://developer.work.weixin.qq.com/document/path/99935) — [wedoc/privilege.go:98](../wedoc/privilege.go#L98)
- [x] 管理收集表
  - [x] [创建收集表](https://developer.work.weixin.qq.com/document/path/97462) — [wedoc/form.go:38](../wedoc/form.go#L38)
  - [x] [编辑收集表](https://developer.work.weixin.qq.com/document/path/97816) — [wedoc/form.go:55](../wedoc/form.go#L55)
  - [x] [获取收集表信息](https://developer.work.weixin.qq.com/document/path/97817) — [wedoc/form.go:71](../wedoc/form.go#L71)
  - [x] [收集表的统计信息查询](https://developer.work.weixin.qq.com/document/path/97818) — [wedoc/form.go:86](../wedoc/form.go#L86)
  - [x] [读取收集表答案](https://developer.work.weixin.qq.com/document/path/97819) — [wedoc/form.go:103](../wedoc/form.go#L103)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/97316)
  - [x] [修改文档成员事件](https://developer.work.weixin.qq.com/document/path/97833) — [wedoc/events.go:51](../wedoc/events.go#L51)
  - [x] [删除文档事件](https://developer.work.weixin.qq.com/document/path/97834) — [wedoc/events.go:51](../wedoc/events.go#L51)
  - [x] [收集表完成事件](https://developer.work.weixin.qq.com/document/path/97835) — [wedoc/events.go:51](../wedoc/events.go#L51)
  - [x] [删除收集表事件](https://developer.work.weixin.qq.com/document/path/98095) — [wedoc/events.go:51](../wedoc/events.go#L51)
  - [x] [修改收集表设置事件](https://developer.work.weixin.qq.com/document/path/98096) — [wedoc/events.go:51](../wedoc/events.go#L51)
  - [x] [字段变更事件](https://developer.work.weixin.qq.com/document/path/100987) — [wedoc/events.go:59](../wedoc/events.go#L59)
  - [x] [记录变更事件](https://developer.work.weixin.qq.com/document/path/100986) — [wedoc/events.go:59](../wedoc/events.go#L59)
- [x] 接收外部数据到智能表格
  - [概述](https://developer.work.weixin.qq.com/document/path/101239)
  - [x] [添加记录](https://developer.work.weixin.qq.com/document/path/101240) — [wedoc/webhook_record.go:32](../wedoc/webhook_record.go#L32)
  - [x] [更新记录](https://developer.work.weixin.qq.com/document/path/101241) — [wedoc/webhook_record.go:39](../wedoc/webhook_record.go#L39)
- [x] 高级功能账号管理
  - [x] [分配高级功能账号](https://developer.work.weixin.qq.com/document/path/99516) — [wedoc/vip.go:28](../wedoc/vip.go#L28)
  - [x] [取消高级功能账号](https://developer.work.weixin.qq.com/document/path/99517) — [wedoc/vip.go:40](../wedoc/vip.go#L40)
  - [x] [获取高级功能账号列表](https://developer.work.weixin.qq.com/document/path/99518) — [wedoc/vip.go:52](../wedoc/vip.go#L52)
- [x] 素材管理
  - [x] [上传文档图片](https://developer.work.weixin.qq.com/document/path/99933) — [wedoc/upload.go:41](../wedoc/upload.go#L41)
  - [x] [上传文档文件](https://developer.work.weixin.qq.com/document/path/101870) — [wedoc/upload.go:58](../wedoc/upload.go#L58)
  - [x] [分块上传文档文件](https://developer.work.weixin.qq.com/document/path/101871) — [wedoc/upload.go:85](../wedoc/upload.go#L85)

### 日程

- [概述](https://developer.work.weixin.qq.com/document/path/93624)
- [x] 管理日历
  - [x] [创建日历](https://developer.work.weixin.qq.com/document/path/93647) — [schedule/schedule.go:64](../schedule/schedule.go#L64)
  - [x] [更新日历](https://developer.work.weixin.qq.com/document/path/97716) — [schedule/schedule.go:130](../schedule/schedule.go#L130)
  - [x] [获取日历详情](https://developer.work.weixin.qq.com/document/path/97717) — [schedule/schedule.go:149](../schedule/schedule.go#L149)
  - [x] [删除日历](https://developer.work.weixin.qq.com/document/path/97718) — [schedule/schedule.go:167](../schedule/schedule.go#L167)
- [x] 管理日程
  - [x] [创建日程](https://developer.work.weixin.qq.com/document/path/93648) — [schedule/schedule.go:220](../schedule/schedule.go#L220)
  - [x] [更新日程](https://developer.work.weixin.qq.com/document/path/97720) — [schedule/schedule.go:324](../schedule/schedule.go#L324)
  - [x] [更新重复日程](https://developer.work.weixin.qq.com/document/path/96204) — [schedule/schedule.go:324](../schedule/schedule.go#L324)
  - [x] [新增日程参与者](https://developer.work.weixin.qq.com/document/path/97721) — [schedule/schedule.go:351](../schedule/schedule.go#L351)
  - [x] [删除日程参与者](https://developer.work.weixin.qq.com/document/path/97722) — [schedule/schedule.go:358](../schedule/schedule.go#L358)
  - [x] [获取日历下的日程列表](https://developer.work.weixin.qq.com/document/path/97723) — [schedule/schedule.go:379](../schedule/schedule.go#L379)
  - [x] [获取日程详情](https://developer.work.weixin.qq.com/document/path/97724) — [schedule/schedule.go:402](../schedule/schedule.go#L402)
  - [x] [取消日程](https://developer.work.weixin.qq.com/document/path/97725) — [schedule/schedule.go:420](../schedule/schedule.go#L420)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/93651)
  - [x] [删除日历事件](https://developer.work.weixin.qq.com/document/path/97728) — [schedule/schedule.go:460](../schedule/schedule.go#L460)
  - [x] [修改日历事件](https://developer.work.weixin.qq.com/document/path/97730) — [schedule/schedule.go:460](../schedule/schedule.go#L460)
  - [x] [修改日程事件](https://developer.work.weixin.qq.com/document/path/97731) — [schedule/schedule.go:460](../schedule/schedule.go#L460)
  - [x] [删除日程事件](https://developer.work.weixin.qq.com/document/path/97732) — [schedule/schedule.go:460](../schedule/schedule.go#L460)
  - [x] [日程回执事件](https://developer.work.weixin.qq.com/document/path/98111) — [schedule/schedule.go:460](../schedule/schedule.go#L460)

### 待办

- [x] [获取待办详情](https://developer.work.weixin.qq.com/document/path/101524) — [todo/todo.go:45](../todo/todo.go#L45)
- [x] [更新待办状态](https://developer.work.weixin.qq.com/document/path/101534) — [todo/todo.go:62](../todo/todo.go#L62)

### 会议

- [概述](https://developer.work.weixin.qq.com/document/path/93626)
- [x] 预约会议基础管理
  - [x] [创建预约会议](https://developer.work.weixin.qq.com/document/path/99104) — [meeting/meeting.go:98](../meeting/meeting.go#L98)
  - [x] [修改预约会议](https://developer.work.weixin.qq.com/document/path/99047) — [meeting/meeting.go:183](../meeting/meeting.go#L183)
  - [x] [取消预约会议](https://developer.work.weixin.qq.com/document/path/99048) — [meeting/meeting.go:206](../meeting/meeting.go#L206)
  - [x] [获取会议详情](https://developer.work.weixin.qq.com/document/path/99049) — [meeting/meeting.go:236](../meeting/meeting.go#L236)
  - [x] [获取成员会议ID列表](https://developer.work.weixin.qq.com/document/path/99050) — [meeting/meeting.go:256](../meeting/meeting.go#L256)
- [x] 会议统计管理
  - [x] [获取会议发起记录](https://developer.work.weixin.qq.com/document/path/99651) — [meeting/meeting.go:328](../meeting/meeting.go#L328)
- [x] 预约会议高级管理
  - [x] [创建预约会议](https://developer.work.weixin.qq.com/document/path/98148) — [meeting/meeting.go:98](../meeting/meeting.go#L98)
  - [x] [修改预约会议](https://developer.work.weixin.qq.com/document/path/98154) — [meeting/meeting.go:183](../meeting/meeting.go#L183)
  - [x] [取消预约会议](https://developer.work.weixin.qq.com/document/path/98153) — [meeting/meeting.go:206](../meeting/meeting.go#L206)
  - [x] [获取会议详情](https://developer.work.weixin.qq.com/document/path/98149) — [meeting/meeting.go:236](../meeting/meeting.go#L236)
  - [x] [获取会议受邀成员列表](https://developer.work.weixin.qq.com/document/path/98160) — [meeting/meeting.go:280](../meeting/meeting.go#L280)
  - [x] [更新会议受邀成员列表](https://developer.work.weixin.qq.com/document/path/98162) — [meeting/meeting.go:303](../meeting/meeting.go#L303)
  - [x] [获取成员会议ID列表](https://developer.work.weixin.qq.com/document/path/98714) — [meeting/meeting.go:256](../meeting/meeting.go#L256)
  - [x] [创建用户专属参会链接](https://developer.work.weixin.qq.com/document/path/98818) — [meeting/meeting.go:352](../meeting/meeting.go#L352)
  - [x] [获取用户专属参会链接](https://developer.work.weixin.qq.com/document/path/98819) — [meeting/meeting.go:373](../meeting/meeting.go#L373)
  - [x] [获取实时会中成员列表](https://developer.work.weixin.qq.com/document/path/98157) — [meeting/attendee.go:36](../meeting/attendee.go#L36)
  - [x] [获取已参会成员列表](https://developer.work.weixin.qq.com/document/path/98156) — [meeting/attendee.go:59](../meeting/attendee.go#L59)
  - [x] [获取实时等候室成员列表](https://developer.work.weixin.qq.com/document/path/98163) — [meeting/attendee.go:103](../meeting/attendee.go#L103)
  - [x] [获取等候室成员记录](https://developer.work.weixin.qq.com/document/path/98164) — [meeting/attendee.go:125](../meeting/attendee.go#L125)
  - [x] [获取成员设备是否入会](https://developer.work.weixin.qq.com/document/path/98165) — [meeting/attendee.go:153](../meeting/attendee.go#L153)
  - [x] [获取会议嘉宾列表](https://developer.work.weixin.qq.com/document/path/99039) — [meeting/attendee.go:184](../meeting/attendee.go#L184)
  - [x] [更新会议嘉宾列表](https://developer.work.weixin.qq.com/document/path/99040) — [meeting/attendee.go:201](../meeting/attendee.go#L201)
  - [x] [获取会议健康度](https://developer.work.weixin.qq.com/document/path/98821) — [meeting/attendee.go:230](../meeting/attendee.go#L230)
  - [x] [修改会议报名配置](https://developer.work.weixin.qq.com/document/path/98797) — [meeting/enroll.go:28](../meeting/enroll.go#L28)
  - [x] [获取会议报名配置](https://developer.work.weixin.qq.com/document/path/98800) — [meeting/enroll.go:43](../meeting/enroll.go#L43)
  - [x] [获取会议成员报名 ID](https://developer.work.weixin.qq.com/document/path/98794) — [meeting/enroll.go:67](../meeting/enroll.go#L67)
  - [x] [获取会议报名信息](https://developer.work.weixin.qq.com/document/path/98810) — [meeting/enroll.go:110](../meeting/enroll.go#L110)
  - [x] [审批会议报名信息](https://developer.work.weixin.qq.com/document/path/98807) — [meeting/enroll.go:133](../meeting/enroll.go#L133)
  - [x] [导入会议报名信息](https://developer.work.weixin.qq.com/document/path/98816) — [meeting/enroll.go:163](../meeting/enroll.go#L163)
  - [x] [删除会议报名信息](https://developer.work.weixin.qq.com/document/path/98817) — [meeting/enroll.go:185](../meeting/enroll.go#L185)
- [x] 会中控制管理
  - [x] [管理会中设置](https://developer.work.weixin.qq.com/document/path/98175) — [meeting/control.go:34](../meeting/control.go#L34)
  - [x] [管理联席主持人](https://developer.work.weixin.qq.com/document/path/98180) — [meeting/control.go:47](../meeting/control.go#L47)
  - [x] [静音成员](https://developer.work.weixin.qq.com/document/path/98184) — [meeting/control.go:65](../meeting/control.go#L65)
  - [x] [关闭或开启成员视频](https://developer.work.weixin.qq.com/document/path/98189) — [meeting/control.go:83](../meeting/control.go#L83)
  - [x] [关闭成员屏幕共享](https://developer.work.weixin.qq.com/document/path/98185) — [meeting/control.go:101](../meeting/control.go#L101)
  - [x] [修改成员在会中显示的昵称](https://developer.work.weixin.qq.com/document/path/98188) — [meeting/control.go:125](../meeting/control.go#L125)
  - [x] [管理等候室成员](https://developer.work.weixin.qq.com/document/path/98186) — [meeting/control.go:150](../meeting/control.go#L150)
  - [x] [移出成员](https://developer.work.weixin.qq.com/document/path/98181) — [meeting/control.go:170](../meeting/control.go#L170)
  - [x] [结束会议](https://developer.work.weixin.qq.com/document/path/98187) — [meeting/control.go:194](../meeting/control.go#L194)
  - [x] [创建会议投票主题](https://developer.work.weixin.qq.com/document/path/98834) — [meeting/poll.go:73](../meeting/poll.go#L73)
  - [x] [修改会议投票主题](https://developer.work.weixin.qq.com/document/path/98835) — [meeting/poll.go:90](../meeting/poll.go#L90)
  - [x] [获取会议投票列表](https://developer.work.weixin.qq.com/document/path/98836) — [meeting/poll.go:103](../meeting/poll.go#L103)
  - [x] [获取会议投票主题信息](https://developer.work.weixin.qq.com/document/path/98837) — [meeting/poll.go:133](../meeting/poll.go#L133)
  - [x] [获取会议投票详情](https://developer.work.weixin.qq.com/document/path/98838) — [meeting/poll.go:154](../meeting/poll.go#L154)
  - [x] [删除会议投票](https://developer.work.weixin.qq.com/document/path/98839) — [meeting/poll.go:175](../meeting/poll.go#L175)
  - [x] [发起会议投票](https://developer.work.weixin.qq.com/document/path/98840) — [meeting/poll.go:195](../meeting/poll.go#L195)
  - [x] [结束会议投票](https://developer.work.weixin.qq.com/document/path/98841) — [meeting/poll.go:218](../meeting/poll.go#L218)
- [x] 网络研讨会 (Webinar) 管理
  - [x] [创建网络研讨会](https://developer.work.weixin.qq.com/document/path/98842) — [meeting/webinar.go:67](../meeting/webinar.go#L67)
  - [x] [修改网络研讨会](https://developer.work.weixin.qq.com/document/path/98843) — [meeting/webinar.go:85](../meeting/webinar.go#L85)
  - [x] [取消网络研讨会](https://developer.work.weixin.qq.com/document/path/98870) — [meeting/webinar.go:98](../meeting/webinar.go#L98)
  - [x] [获取网络研讨会详情](https://developer.work.weixin.qq.com/document/path/98860) — [meeting/webinar.go:113](../meeting/webinar.go#L113)
  - [x] [获取网络研讨会嘉宾列表](https://developer.work.weixin.qq.com/document/path/98871) — [meeting/webinar.go:130](../meeting/webinar.go#L130)
  - [x] [更新网络研讨会嘉宾列表](https://developer.work.weixin.qq.com/document/path/98872) — [meeting/webinar.go:151](../meeting/webinar.go#L151)
  - [x] [管理网络研讨会暖场配置](https://developer.work.weixin.qq.com/document/path/98882) — [meeting/webinar.go:176](../meeting/webinar.go#L176)
  - [x] [修改网络研讨会报名配置](https://developer.work.weixin.qq.com/document/path/98875) — [meeting/webinar.go:198](../meeting/webinar.go#L198)
  - [x] [获取网络研讨会报名配置](https://developer.work.weixin.qq.com/document/path/98874) — [meeting/webinar.go:214](../meeting/webinar.go#L214)
  - [x] [获取网络研讨会成员报名 ID](https://developer.work.weixin.qq.com/document/path/98873) — [meeting/webinar.go:232](../meeting/webinar.go#L232)
  - [x] [获取网络研讨会报名信息](https://developer.work.weixin.qq.com/document/path/98876) — [meeting/webinar.go:255](../meeting/webinar.go#L255)
  - [x] [审批网络研讨会报名信息](https://developer.work.weixin.qq.com/document/path/98877) — [meeting/webinar.go:277](../meeting/webinar.go#L277)
  - [x] [导入网络研讨会报名信息](https://developer.work.weixin.qq.com/document/path/98880) — [meeting/webinar.go:300](../meeting/webinar.go#L300)
  - [x] [删除网络研讨会报名信息](https://developer.work.weixin.qq.com/document/path/98881) — [meeting/webinar.go:323](../meeting/webinar.go#L323)
- [x] 电话入会（PSTN）管理
  - [x] [批量外呼](https://developer.work.weixin.qq.com/document/path/98823) — [meeting/phone.go:35](../meeting/phone.go#L35)
  - [x] [获取会议的外呼状态](https://developer.work.weixin.qq.com/document/path/98824) — [meeting/phone.go:54](../meeting/phone.go#L54)
  - [x] [获取电话入会的成员ID](https://developer.work.weixin.qq.com/document/path/98825) — [meeting/phone.go:76](../meeting/phone.go#L76)
- [x] Rooms会议室管理
  - [x] [预定Rooms会议室](https://developer.work.weixin.qq.com/document/path/98791) — [meeting/rooms.go:96](../meeting/rooms.go#L96)
  - [x] [释放Rooms会议室](https://developer.work.weixin.qq.com/document/path/98792) — [meeting/rooms.go:118](../meeting/rooms.go#L118)
  - [x] [获取Rooms会议室列表](https://developer.work.weixin.qq.com/document/path/98795) — [meeting/rooms.go:135](../meeting/rooms.go#L135)
  - [x] [获取Rooms会议室详情](https://developer.work.weixin.qq.com/document/path/98793) — [meeting/rooms.go:153](../meeting/rooms.go#L153)
  - [x] [获取Rooms会议室配置项](https://developer.work.weixin.qq.com/document/path/98802) — [meeting/rooms.go:170](../meeting/rooms.go#L170)
  - [x] [获取Rooms会议室下的会议列表](https://developer.work.weixin.qq.com/document/path/98796) — [meeting/rooms.go:187](../meeting/rooms.go#L187)
  - [x] [获取设备列表](https://developer.work.weixin.qq.com/document/path/98798) — [meeting/phone.go:117](../meeting/phone.go#L117)
  - [x] [获取控制器列表](https://developer.work.weixin.qq.com/document/path/98799) — [meeting/phone.go:152](../meeting/phone.go#L152)
  - [x] [获取Rooms会议室资源](https://developer.work.weixin.qq.com/document/path/98809) — [meeting/rooms.go:213](../meeting/rooms.go#L213)
  - [x] [呼叫Rooms会议室](https://developer.work.weixin.qq.com/document/path/98804) — [meeting/rooms.go:225](../meeting/rooms.go#L225)
  - [x] [取消呼叫Rooms会议室](https://developer.work.weixin.qq.com/document/path/98805) — [meeting/rooms.go:247](../meeting/rooms.go#L247)
  - [x] [获取Rooms会议室应答状态](https://developer.work.weixin.qq.com/document/path/98806) — [meeting/rooms.go:272](../meeting/rooms.go#L272)
- [x] 会议室连接器（MRA）管理
  - [x] [获取 MRA 状态信息](https://developer.work.weixin.qq.com/document/path/98786) — [meeting/rooms.go:307](../meeting/rooms.go#L307)
  - [x] [切换 MRA 默认布局](https://developer.work.weixin.qq.com/document/path/98787) — [meeting/rooms.go:326](../meeting/rooms.go#L326)
  - [x] [设置 MRA 举手或手放下](https://developer.work.weixin.qq.com/document/path/98788) — [meeting/rooms.go:348](../meeting/rooms.go#L348)
  - [x] [挂断 MRA 呼叫](https://developer.work.weixin.qq.com/document/path/98789) — [meeting/rooms.go:369](../meeting/rooms.go#L369)
- [x] 会议布局和背景管理
  - [x] [获取布局模板列表](https://developer.work.weixin.qq.com/document/path/98844) — [meeting/layout.go:60](../meeting/layout.go#L60)
  - [x] [添加会议基础布局](https://developer.work.weixin.qq.com/document/path/98845) — [meeting/layout.go:75](../meeting/layout.go#L75)
  - [x] [添加会议高级布局](https://developer.work.weixin.qq.com/document/path/98861) — [meeting/layout.go:97](../meeting/layout.go#L97)
  - [x] [修改会议基础布局](https://developer.work.weixin.qq.com/document/path/98846) — [meeting/layout.go:117](../meeting/layout.go#L117)
  - [x] [修改会议高级布局](https://developer.work.weixin.qq.com/document/path/98868) — [meeting/layout.go:136](../meeting/layout.go#L136)
  - [x] [设置会议默认布局](https://developer.work.weixin.qq.com/document/path/98847) — [meeting/layout.go:156](../meeting/layout.go#L156)
  - [x] [设置高级布局](https://developer.work.weixin.qq.com/document/path/98869) — [meeting/layout.go:173](../meeting/layout.go#L173)
  - [x] [获取会议布局列表](https://developer.work.weixin.qq.com/document/path/98862) — [meeting/layout.go:196](../meeting/layout.go#L196)
  - [x] [获取用户布局](https://developer.work.weixin.qq.com/document/path/98865) — [meeting/layout.go:213](../meeting/layout.go#L213)
  - [x] [批量删除布局](https://developer.work.weixin.qq.com/document/path/98866) — [meeting/layout.go:242](../meeting/layout.go#L242)
  - [x] [添加会议背景](https://developer.work.weixin.qq.com/document/path/98851) — [meeting/layout.go:278](../meeting/layout.go#L278)
  - [x] [设置会议默认背景](https://developer.work.weixin.qq.com/document/path/98852) — [meeting/layout.go:298](../meeting/layout.go#L298)
  - [x] [获取会议背景列表](https://developer.work.weixin.qq.com/document/path/98856) — [meeting/layout.go:316](../meeting/layout.go#L316)
  - [x] [删除会议背景](https://developer.work.weixin.qq.com/document/path/98853) — [meeting/layout.go:333](../meeting/layout.go#L333)
  - [x] [批量删除会议背景](https://developer.work.weixin.qq.com/document/path/98854) — [meeting/layout.go:351](../meeting/layout.go#L351)
- [x] 录制管理
  - [x] [获取会议录制列表](https://developer.work.weixin.qq.com/document/path/98192) — [meeting/record.go:114](../meeting/record.go#L114)
  - [x] [获取录制文件访问统计](https://developer.work.weixin.qq.com/document/path/98209) — [meeting/record.go:138](../meeting/record.go#L138)
  - [x] [修改会议录制共享设置](https://developer.work.weixin.qq.com/document/path/98208) — [meeting/record.go:161](../meeting/record.go#L161)
  - [x] [删除会议录制](https://developer.work.weixin.qq.com/document/path/98206) — [meeting/record.go:180](../meeting/record.go#L180)
  - [x] [删除单个录制文件](https://developer.work.weixin.qq.com/document/path/98207) — [meeting/record.go:197](../meeting/record.go#L197)
  - [x] [获取单个录制文件详情](https://developer.work.weixin.qq.com/document/path/98205) — [meeting/record.go:229](../meeting/record.go#L229)
  - [x] [获取会议录制地址](https://developer.work.weixin.qq.com/document/path/98196) — [meeting/record.go:248](../meeting/record.go#L248)
  - [x] [获取录制转写段落信息](https://developer.work.weixin.qq.com/document/path/98212) — [meeting/record.go:267](../meeting/record.go#L267)
  - [x] [获取录制转写详情](https://developer.work.weixin.qq.com/document/path/98211) — [meeting/record.go:290](../meeting/record.go#L290)
  - [x] [获取录制转写搜索结果](https://developer.work.weixin.qq.com/document/path/98213) — [meeting/record.go:317](../meeting/record.go#L317)
- [x] 高级功能账号管理
  - [x] [分配高级功能账号](https://developer.work.weixin.qq.com/document/path/99508) — [meeting/phone.go:188](../meeting/phone.go#L188)
  - [x] [取消高级功能账号](https://developer.work.weixin.qq.com/document/path/99509) — [meeting/phone.go:211](../meeting/phone.go#L211)
  - [x] [获取高级功能账号列表](https://developer.work.weixin.qq.com/document/path/99510) — [meeting/phone.go:247](../meeting/phone.go#L247)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/99103)
  - [x] [修改会议事件](https://developer.work.weixin.qq.com/document/path/99081) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [取消会议事件](https://developer.work.weixin.qq.com/document/path/99082) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议开始事件](https://developer.work.weixin.qq.com/document/path/98333) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议结束事件](https://developer.work.weixin.qq.com/document/path/98337) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议全体静音事件](https://developer.work.weixin.qq.com/document/path/98341) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议解除全体静音事件](https://developer.work.weixin.qq.com/document/path/98345) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员入会事件](https://developer.work.weixin.qq.com/document/path/98348) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员离会事件](https://developer.work.weixin.qq.com/document/path/98352) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员等待主持人入会事件](https://developer.work.weixin.qq.com/document/path/98353) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员进入等候室事件](https://developer.work.weixin.qq.com/document/path/98354) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员离开等候室事件](https://developer.work.weixin.qq.com/document/path/98355) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员从等候室进入会议事件](https://developer.work.weixin.qq.com/document/path/98393) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [成员从会议中被移入等候室事件](https://developer.work.weixin.qq.com/document/path/98394) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [共享屏幕开启事件](https://developer.work.weixin.qq.com/document/path/98395) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [共享屏幕结束事件](https://developer.work.weixin.qq.com/document/path/98396) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议成员角色变更事件](https://developer.work.weixin.qq.com/document/path/98397) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [网络研讨会成员角色变更事件](https://developer.work.weixin.qq.com/document/path/98771) — [meeting/webinar.go:386](../meeting/webinar.go#L386)
  - [x] [网络研讨会暖场上传结果](https://developer.work.weixin.qq.com/document/path/98773) — [meeting/webinar.go:386](../meeting/webinar.go#L386)
  - [x] [PSTN 外呼状态更新事件](https://developer.work.weixin.qq.com/document/path/98774) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [素材上传结果](https://developer.work.weixin.qq.com/document/path/98775) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [开始云录制事件](https://developer.work.weixin.qq.com/document/path/98398) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [暂停云录制事件](https://developer.work.weixin.qq.com/document/path/98399) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [恢复云录制事件](https://developer.work.weixin.qq.com/document/path/98400) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [停止云录制事件](https://developer.work.weixin.qq.com/document/path/98401) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [云录制已完成事件](https://developer.work.weixin.qq.com/document/path/98402) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [删除云录制事件](https://developer.work.weixin.qq.com/document/path/98404) — [meeting/record.go:369](../meeting/record.go#L369)
  - [x] [用户报名事件](https://developer.work.weixin.qq.com/document/path/98781) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [用户取消报名事件](https://developer.work.weixin.qq.com/document/path/98782) — [meeting/event.go:105](../meeting/event.go#L105)
  - [x] [会议室应答事件](https://developer.work.weixin.qq.com/document/path/98783) — [meeting/rooms.go:404](../meeting/rooms.go#L404)
  - [x] [会议发起事件](https://developer.work.weixin.qq.com/document/path/99648) — [meeting/event.go:137](../meeting/event.go#L137)

### 微盘

- [概述](https://developer.work.weixin.qq.com/document/path/93654)
- [x] 管理空间
  - [x] [新建空间](https://developer.work.weixin.qq.com/document/path/93655) — [wedrive/space.go:82](../wedrive/space.go#L82)
  - [x] [重命名空间](https://developer.work.weixin.qq.com/document/path/97856) — [wedrive/space.go:105](../wedrive/space.go#L105)
  - [x] [解散空间](https://developer.work.weixin.qq.com/document/path/97857) — [wedrive/space.go:122](../wedrive/space.go#L122)
  - [x] [获取空间信息](https://developer.work.weixin.qq.com/document/path/97858) — [wedrive/space.go:137](../wedrive/space.go#L137)
- [x] 管理空间权限
  - [x] [添加成员/部门](https://developer.work.weixin.qq.com/document/path/93656) — [wedrive/space.go:176](../wedrive/space.go#L176)
  - [x] [移除成员/部门](https://developer.work.weixin.qq.com/document/path/97875) — [wedrive/space.go:198](../wedrive/space.go#L198)
  - [x] [安全设置](https://developer.work.weixin.qq.com/document/path/97876) — [wedrive/space.go:218](../wedrive/space.go#L218)
  - [x] [获取邀请链接](https://developer.work.weixin.qq.com/document/path/97877) — [wedrive/space.go:231](../wedrive/space.go#L231)
  - [x] [获取空间信息](https://developer.work.weixin.qq.com/document/path/97878) — [wedrive/space.go:157](../wedrive/space.go#L157)
- [x] 管理文件
  - [x] [获取文件列表](https://developer.work.weixin.qq.com/document/path/93657) — [wedrive/file.go:69](../wedrive/file.go#L69)
  - [x] [上传文件](https://developer.work.weixin.qq.com/document/path/97880) — [wedrive/file.go:106](../wedrive/file.go#L106)
  - [x] [文件分块上传](https://developer.work.weixin.qq.com/document/path/98004) — [wedrive/file.go:158](../wedrive/file.go#L158)、[wedrive/upload.go:24](../wedrive/upload.go#L24)
  - [x] [下载文件](https://developer.work.weixin.qq.com/document/path/97881) — [wedrive/file.go:241](../wedrive/file.go#L241)
  - [x] [新建文件夹/文档](https://developer.work.weixin.qq.com/document/path/97882) — [wedrive/file.go:260](../wedrive/file.go#L260)
  - [x] [重命名文件](https://developer.work.weixin.qq.com/document/path/97883) — [wedrive/file.go:284](../wedrive/file.go#L284)
  - [x] [移动文件](https://developer.work.weixin.qq.com/document/path/97884) — [wedrive/file.go:305](../wedrive/file.go#L305)
  - [x] [删除文件](https://developer.work.weixin.qq.com/document/path/97885) — [wedrive/file.go:327](../wedrive/file.go#L327)
  - [x] [获取文件信息](https://developer.work.weixin.qq.com/document/path/97886) — [wedrive/file.go:341](../wedrive/file.go#L341)
- [x] 管理文件权限
  - [x] [新增成员](https://developer.work.weixin.qq.com/document/path/93658) — [wedrive/file.go:360](../wedrive/file.go#L360)
  - [x] [删除成员](https://developer.work.weixin.qq.com/document/path/97888) — [wedrive/file.go:380](../wedrive/file.go#L380)
  - [x] [分享设置](https://developer.work.weixin.qq.com/document/path/97889) — [wedrive/file.go:400](../wedrive/file.go#L400)
  - [x] [获取分享链接](https://developer.work.weixin.qq.com/document/path/97890) — [wedrive/file.go:423](../wedrive/file.go#L423)
  - [x] [获取文件权限信息](https://developer.work.weixin.qq.com/document/path/97891) — [wedrive/file.go:479](../wedrive/file.go#L479)
  - [x] [修改文件安全设置](https://developer.work.weixin.qq.com/document/path/97892) — [wedrive/file.go:507](../wedrive/file.go#L507)
- [x] 回调通知
  - [概述](https://developer.work.weixin.qq.com/document/path/97482)
  - [x] [微盘容量不足事件](https://developer.work.weixin.qq.com/document/path/97898) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
  - [x] [空间变更事件](https://developer.work.weixin.qq.com/document/path/97899) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
  - [x] [文件变更事件](https://developer.work.weixin.qq.com/document/path/97900) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
  - [x] [解散空间](https://developer.work.weixin.qq.com/document/path/97901) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
  - [x] [修改空间成员](https://developer.work.weixin.qq.com/document/path/97902) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
  - [x] [修改空间安全设置](https://developer.work.weixin.qq.com/document/path/97903) — [wedrive/callback.go:38](../wedrive/callback.go#L38)
- [x] 高级功能账号管理
  - [x] [分配高级功能账号](https://developer.work.weixin.qq.com/document/path/99512) — [wedrive/vip.go:18](../wedrive/vip.go#L18)
  - [x] [取消高级功能账号](https://developer.work.weixin.qq.com/document/path/99513) — [wedrive/vip.go:32](../wedrive/vip.go#L32)
  - [x] [获取高级功能账号列表](https://developer.work.weixin.qq.com/document/path/99514) — [wedrive/vip.go:46](../wedrive/vip.go#L46)
- [x] [版本和容量管理](https://developer.work.weixin.qq.com/document/path/95856) — [wedrive/vip.go:76](../wedrive/vip.go#L76)

### 直播

- [概述](https://developer.work.weixin.qq.com/document/path/93633)
- [x] [创建预约直播](https://developer.work.weixin.qq.com/document/path/93637) — [live/live.go:110](../live/live.go#L110)
- [x] [修改预约直播](https://developer.work.weixin.qq.com/document/path/93640) — [live/live.go:130](../live/live.go#L130)
- [x] [取消预约直播](https://developer.work.weixin.qq.com/document/path/93638) — [live/live.go:143](../live/live.go#L143)
- [x] [删除直播回放](https://developer.work.weixin.qq.com/document/path/93874) — [live/live.go:158](../live/live.go#L158)
- [x] [在微信中观看直播或直播回放](https://developer.work.weixin.qq.com/document/path/93641) — [live/live.go:173](../live/live.go#L173)
- [x] [获取成员直播ID列表](https://developer.work.weixin.qq.com/document/path/93634) — [live/live.go:200](../live/live.go#L200)
- [x] [获取直播详情](https://developer.work.weixin.qq.com/document/path/93635) — [live/live.go:221](../live/live.go#L221)
- [x] [获取直播观看明细](https://developer.work.weixin.qq.com/document/path/93636) — [live/live.go:240](../live/live.go#L240)
- [x] [直播回调事件](https://developer.work.weixin.qq.com/document/path/94145) — [live/live.go:311](../live/live.go#L311)
- [x] [获取跳转小程序商城的直播观众信息](https://developer.work.weixin.qq.com/document/path/94442) — [live/live.go:281](../live/live.go#L281)

### 公费电话

- [x] [获取公费电话拨打记录](https://developer.work.weixin.qq.com/document/path/93662) — [dial/dial.go:41](../dial/dial.go#L41)

### 打卡

- [x] [获取企业所有打卡规则](https://developer.work.weixin.qq.com/document/path/93384) — [checkin/checkin.go:265](../checkin/checkin.go#L265)
- [x] [获取员工打卡规则](https://developer.work.weixin.qq.com/document/path/90263) — [checkin/checkin.go:279](../checkin/checkin.go#L279)
- [x] [获取打卡记录数据](https://developer.work.weixin.qq.com/document/path/90262) — [checkin/checkin.go:89](../checkin/checkin.go#L89)
- [x] [获取打卡日报数据](https://developer.work.weixin.qq.com/document/path/93374) — [checkin/checkin.go:137](../checkin/checkin.go#L137)
- [x] [获取打卡月报数据](https://developer.work.weixin.qq.com/document/path/93387) — [checkin/checkin.go:144](../checkin/checkin.go#L144)
- [x] [获取打卡人员排班信息](https://developer.work.weixin.qq.com/document/path/93380) — [checkin/checkin.go:171](../checkin/checkin.go#L171)
- [x] [为打卡人员排班](https://developer.work.weixin.qq.com/document/path/93385) — [checkin/checkin.go:193](../checkin/checkin.go#L193)
- [x] [为打卡人员补卡](https://developer.work.weixin.qq.com/document/path/95803) — [checkin/checkin.go:211](../checkin/checkin.go#L211)
- [x] [添加打卡记录](https://developer.work.weixin.qq.com/document/path/99647) — [checkin/checkin.go:225](../checkin/checkin.go#L225)
- [x] [录入打卡人员人脸信息](https://developer.work.weixin.qq.com/document/path/93378) — [checkin/checkin.go:247](../checkin/checkin.go#L247)
- [x] [获取设备打卡数据](https://developer.work.weixin.qq.com/document/path/94126) — [checkin/checkin.go:113](../checkin/checkin.go#L113)
- [x] [管理打卡规则](https://developer.work.weixin.qq.com/document/path/98041) — [checkin/checkin.go:301](../checkin/checkin.go#L301)

### 审批

- [概述](https://developer.work.weixin.qq.com/document/path/91854)
- [x] [获取审批模板详情](https://developer.work.weixin.qq.com/document/path/91982) — [approval/approval.go:108](../approval/approval.go#L108)
- [x] [提交审批申请](https://developer.work.weixin.qq.com/document/path/91853) — [approval/approval.go:125](../approval/approval.go#L125)
- [x] [审批申请状态变化回调通知](https://developer.work.weixin.qq.com/document/path/91815) — [approval/approval.go:362](../approval/approval.go#L362)
- [x] [批量获取审批单号](https://developer.work.weixin.qq.com/document/path/91816) — [approval/approval.go:145](../approval/approval.go#L145)
- [x] [获取审批申请详情](https://developer.work.weixin.qq.com/document/path/91983) — [approval/approval.go:163](../approval/approval.go#L163)
- [x] [获取审批数据（旧）](https://developer.work.weixin.qq.com/document/path/91530) — [approval/approval.go:183](../approval/approval.go#L183)
- [x] [获取企业假期管理配置](https://developer.work.weixin.qq.com/document/path/93375) — [approval/approval.go:228](../approval/approval.go#L228)
- [x] [获取成员假期余额](https://developer.work.weixin.qq.com/document/path/93376) — [approval/approval.go:242](../approval/approval.go#L242)
- [x] [修改成员假期余额](https://developer.work.weixin.qq.com/document/path/93377) — [approval/approval.go:261](../approval/approval.go#L261)
- [x] [审批流程引擎](https://developer.work.weixin.qq.com/document/path/90269) — [approval/approval.go:208](../approval/approval.go#L208)
- [x] [创建审批模板](https://developer.work.weixin.qq.com/document/path/97437) — [approval/approval.go:275](../approval/approval.go#L275)
- [x] [更新审批模板](https://developer.work.weixin.qq.com/document/path/97438) — [approval/approval.go:293](../approval/approval.go#L293)

### 汇报

- [概述](https://developer.work.weixin.qq.com/document/path/93496)
- [x] [批量获取汇报记录单号](https://developer.work.weixin.qq.com/document/path/93393) — [journal/journal.go:61](../journal/journal.go#L61)
- [x] [获取汇报记录详情](https://developer.work.weixin.qq.com/document/path/93394) — [journal/journal.go:76](../journal/journal.go#L76)
- [x] [获取汇报统计数据](https://developer.work.weixin.qq.com/document/path/93395) — [journal/journal.go:102](../journal/journal.go#L102)
- [x] [下载微盘文件](https://developer.work.weixin.qq.com/document/path/98021) — [journal/journal.go:135](../journal/journal.go#L135)

### 人事助手

- [x] 花名册
  - [概述](https://developer.work.weixin.qq.com/document/path/99130)
  - [x] [获取员工字段配置](https://developer.work.weixin.qq.com/document/path/99131) — [hr/hr.go:39](../hr/hr.go#L39)
  - [x] [获取员工花名册信息](https://developer.work.weixin.qq.com/document/path/99132) — [hr/hr.go:61](../hr/hr.go#L61)
  - [x] [更新员工花名册信息](https://developer.work.weixin.qq.com/document/path/99133) — [hr/hr.go:103](../hr/hr.go#L103)

### 会议室

- [概述](https://developer.work.weixin.qq.com/document/path/93618)
- [x] [会议室管理](https://developer.work.weixin.qq.com/document/path/93619) — [meetingroom/meetingroom.go:82](../meetingroom/meetingroom.go#L82)
- [x] [会议室预定管理](https://developer.work.weixin.qq.com/document/path/93620) — [meetingroom/meetingroom.go:140](../meetingroom/meetingroom.go#L140)
- [x] [回调事件](https://developer.work.weixin.qq.com/document/path/95333) — [meetingroom/meetingroom.go:278](../meetingroom/meetingroom.go#L278)

### 高级功能

- [概述](https://developer.work.weixin.qq.com/document/path/99860)
- [x] [成员申请的提交回调](https://developer.work.weixin.qq.com/document/path/99876) — [advanced/advanced.go:149](../advanced/advanced.go#L149)
- [x] [成员申请的终止回调](https://developer.work.weixin.qq.com/document/path/99877) — [advanced/advanced.go:166](../advanced/advanced.go#L166)
- [x] [设置审批单审批信息](https://developer.work.weixin.qq.com/document/path/99880) — [advanced/advanced.go:75](../advanced/advanced.go#L75)
- [x] [批量获取申请单ID](https://developer.work.weixin.qq.com/document/path/99883) — [advanced/advanced.go:89](../advanced/advanced.go#L89)
- [x] [获取申请单详细信息](https://developer.work.weixin.qq.com/document/path/99885) — [advanced/advanced.go:105](../advanced/advanced.go#L105)

### 紧急通知应用

- [概述](https://developer.work.weixin.qq.com/document/path/91623)
- [x] [发起语音电话](https://developer.work.weixin.qq.com/document/path/91627) — [pstn/pstn.go:29](../pstn/pstn.go#L29)
- [x] [获取接听状态](https://developer.work.weixin.qq.com/document/path/91628) — [pstn/pstn.go:57](../pstn/pstn.go#L57)
