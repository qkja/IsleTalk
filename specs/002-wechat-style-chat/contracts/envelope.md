# 契约：统一信封与分页/错误语义（对外接口）

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

本文是**客户端 ↔ 后端**（对外北向接口）的统一契约。落地宪法原则 III 与 B 组。

## 统一信封

每个对外接口返回同一信封，顶层字段集合一致，禁止个别接口自定义顶层字段：

```json
{
  "success": true,
  "data": { },
  "message": null,
  "meta": null
}
```

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `success` | bool | 是 | 成功标识 |
| `data` | object \| null | 是 | 业务数据；失败时为 `null` |
| `message` | string \| null | 是 | 面向用户的错误/提示文案；成功时 `null` |
| `meta` | object \| null | 否 | 分页与元信息；分页接口必填 |

- 传输层成功与业务成功分离：HTTP 200 不代表业务成功，业务失败 MUST `success=false` + 稳定错误码。
- `message` 是**面向用户的安全文案**（宪法 E 组），不得含堆栈、查询语句、密钥、内部路径/主机名；
  机器可读的错误码在 `meta.code`（或 `data` 外层的稳定 `code` 字段，见 error-registry.md）。
- 命名 camelCase；时间 ISO 8601 UTC；敏感响应带 `Cache-Control: no-store`。

## 分页（AIP-158）

列表/查询接口 MUST 分页，禁止无界返回。

- 入参：`page_size`（默认 50、最大 1000、超上限收敛到 1000 而非报错、负值返回 `INVALID_ARGUMENT`）、
  `page_token`（不透明字符串，禁止 offset）。
- 返回：`meta = { total, page, limit, next_page_token }`；`total` 为总数，`next_page_token` 为空表示末页。
- 翻页期间排序 MUST 稳定；禁止"先取全量再切片"。

## 错误语义（AIP-193 / RFC 9457）

- 统一错误结构：`code`（稳定机器码）、`message`（简短可执行）、`details`（可选，独立字段承载上下文）。
- 客户端 MUST NOT 解析 `message` 文本做分支，一律用 `code`。
- 未授权响应 MUST NOT 区分"对象不存在"与"无权访问"（API1:2023 防枚举）。
- 状态码语义正确：禁止用 200 包装所有错误；错误码是否可重试见 error-registry.md。

## 幂等（AIP-192）

- 对外写接口 SHOULD 支持 `Idempotency-Key` 头，服务端保留 24 小时；重复提交返回首次结果。
- 消息发送以消息 ID 幂等去重（FR-028），与 `Idempotency-Key` 双层保障。

## 版本与下线（RFC 8594）

- 对外接口 MUST 明确版本策略；废弃 MUST 提前 ≥90 天公告，经 `Sunset` 头标记下线时间，禁止无限期双版本。
