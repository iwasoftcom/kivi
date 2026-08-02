Kivi

会出示回执的数据库 · v1.1.0 · 格式 v2

一个用纯 Go 编写的事件账本数据库——单一静态二进制文件，能为它给出的每一个答案负责。

## Kivi 是什么

Kivi 只存储**事件**——不可变、哈希链接的记录。任何看起来像当前状态的东西 （一张表、一个图、一条时间序列、一个向量索引）都是**视图**：对这些事件的一次折叠， 按需重新编译，随时可丢弃。每个答案都带有**溯源**（确立它的记录编号）和 **范围**（它回看了多远的历史），而 `why` 会返回回执记录本身。 缺失的数据是诚实的拒绝，绝不是编造。

## 三条原则

**只有事件是持久的。**  
没有 UPDATE。新事实被追加； 过去从不被改写。历史是一等公民。

**表示是被编译的。**  
删除 `derived/` 永远合法： 答案不变，只是下一次查询更慢。

**没有溯源就没有答案。**  
无溯源的答案无法表示—— 在内核和每个 SDK 中皆然。没有任何东西是编造的。

## 它与普通数据库的区别

|  | 典型数据库 | Kivi |
| --- | --- | --- |
| 什么是持久的 | 当前状态；UPDATE 摧毁过去 | 事件；状态由其重新编译 |
| 答案 | 赤裸的值 | 值 + 溯源 + 范围；`why` 取回回执 |
| 时间点 | 快照与额外机制 | 设计上免费：折叠到记录 N（`--as-of`） |
| 完整性 | 假定 | 已验证：SHA-256 链 + Ed25519 封印；客户端重新验证服务器 |
| 删除 | 行没了，故事也没了 | 加密擦除：密钥销毁、字节不变、擦除本身留有回执 |

## 快速开始

运行服务器（单一静态二进制文件，或容器镜像）：

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# 或使用二进制文件
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

追加一个事件并读回——答案会带着它的溯源到来：

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # 重新哈希整条链，校验每一个封印
kivi why    ./kivi.ledger 0      # 答案背后的回执
```

客户端以**六种语言**遵循同一契约——Go、Python、 Java & Kotlin、Rust、Node.js 和 .NET——提供类型化的 “实体进、实体出”访问，并默认开启客户端验证。

## 内含什么

**可溯源的答案**  
每个值都引用确立它的事件； 无溯源的答案无法构造。

**时间旅行**  
“我们在记录 N 时知道什么？”是一次有界回放—— 无快照、无迁移（`--as-of`）。

**已验证的完整性**  
SHA-256 哈希链 + Ed25519 封印； 哪怕单个字节被篡改也会被抓出并指名。

**加密擦除**  
每条记录一把密钥；擦除会销毁密钥、 保持链有效，并为擦除本身留下回执。

**ACID（仅追加形态）**  
逐事件原子性、串行化写入、 快照读取、组提交 `fsync` 持久性。

**集群与联邦**  
多数派提交的故障转移、只读副本， 以及独立账本之间的相互见证。

**管理面板与身份**  
内嵌 UI、留有回执的配置、 用户/角色/会话，以及具名、可吊销的 API 密钥。

**LLM 门（MCP）**  
为 AI 智能体提供一段会引用来源的记忆； 缺失的事实会被拒绝，而非臆造。

**多租户，按需开通**  
一个进程中运行多个彼此隔离的租户——新租户可在运行时通过管理 API 开通，留有回执，无需重启。

**以服务方式运行**  
Linux 上用 systemd（.deb/.rpm），Windows 上由安装程序注册为真正的 Windows 服务——开机自启，停止时优雅收尾。

## 管理面板

每个服务器都内置一个 Web 管理面板（无需单独部署），位于 ops 端口。打开 `https://<host>:4742/admin/` 并用一个用户账户登录——在那里， 实时且无需触及数据面，你可以：

-   **概览与周期**——健康状况、记录与封印计数、轮换。
-   **记录与搜索**——追加的实时流，外加语义搜索。
-   **集群与联邦**——拓扑、故障转移与见证状态。
-   **用户与 API 密钥**——创建账户和角色，签发与吊销密钥。
-   **配置**——重新调整限额与设置；每次更改都会留有回执写入账本（密钥绝不入账）。

## 架构

一个进程、两道网络门、一条严格的依赖规则：内核只使用 Go 标准库； gRPC 位于边缘。磁盘上恰好只有一种持久真相——事件日志——其余一切都可丢弃。

CLI · 六语言 SDK 不轻信：在客户端验证 管理面板 · LLM 智能体（MCP） 内嵌 UI · 会引用来源的记忆 gRPC 数据面 · :4741 ops HTTP · :4742 内核（Go 标准库）：账本 · 视图 · 溯源 · 封印 · 身份 答案在查询时编译——任何派生物都不是主记录 磁盘：仅追加的事件日志（+ 封印）

## 客户端 SDK

六个客户端 SDK 说着同一套线路契约——全部采用 MIT 许可，并发布到各自语言的 官方仓库。kivi 服务端与内核保持专有；客户端是开源的。

| 语言 | 仓库 | 安装 | 包页面 |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### 不轻信的客户端

每个 SDK 都暴露同样精简的类型化接口——而且客户端验证**默认开启**：客户端 在读取时会重新哈希整条链并核对 Ed25519 封印，因此撒谎的服务端或哪怕一个字节 被翻转都会被抓住，而不是被信任。

-   `append(type, body)`——追加一个事件，获得它的回执。
-   `table(subject, attribute)`——一个**带溯源的答案**： `value` + `trace`（确立该答案的记录编号） + `scope`。缺失的单元格是一次诚实拒绝——一个原生异常， 绝不是伪造的 null。
-   `table(…, as_of=N)`——同一个答案**截至记录 N 时**的状态 （时间旅行）。
-   `why(trace)`——回执记录本身。
-   `replay()`——流式返回每一条记录，并在客户端重新验证。
-   `similar(query, k)`——带溯源的语义搜索（记录 + 分数）。
-   `login(user, pw)`——一个按角色限定范围的会话令牌； `with_bearer` 以另一身份复用同一通道。
-   `head()`——低成本的定位（头部编号 + 哈希），不做审计。

### 示例（Python）

在全部六种语言中形态完全一致：

```
from kivi import KiviClient

c = KiviClient("localhost:4741", token="…")     # verify=True by default
c.append("property", {"subject": "dog", "attribute": "sound", "value": "bark"})
a = c.table("dog", "sound")              # TracedAnswer(value="bark", trace=[0], scope=0)
old = c.table("dog", "sound", as_of=41)  # the answer as of record 41
receipts = c.why(a.trace)                # the actual ledger records
for rec in c.replay():                   # hash + chain + seal verified CLIENT-SIDE
    ...
```

一套一致性测试对全部六种语言运行，因此每种语言都返回相同的答案与相同的 溯源。各语言的示例（Go、Java / Kotlin、Rust、Node.js、.NET）以及 类型化实体层，都在[完整参考（英文）](reference.en.html)中。

## 完整文档与源代码

-   **[完整参考（英文）](reference.en.html)**——每一项功能、 CLI 与 gRPC API、环境变量、操作指南，以及按语言分页的代码示例。
-   **兼容性：**gRPC API、磁盘格式 v2 与 SDK 表面是一份 SemVer 契约—— 不经主版本号跃迁不会破坏。
-   **诚实的状态：**尚未经过独立安全审计；尚无生产里程。这些是披露， 而非对稳定性承诺的保留。

## 联系我们

问题、演示，或成为设计合作伙伴的洽谈——联系我们：

-   **邮箱** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[联系我们](#contact) · Kivi v1.1.0 · 磁盘格式 v2（冻结契约、黄金向量）· Go 标准库内核， 边缘用 gRPC · 单一静态二进制文件。© iwasoft。
