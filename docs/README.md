# 文档索引（AI）

> **强制规则**：[`.cursor/rules/docs-reading-order.mdc`](../.cursor/rules/docs-reading-order.mdc)

## 阅读顺序

| 步 | 文件 | 必读 |
|----|------|------|
| **1** | [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) | **全文**：30 条 REQ、顺序、实现要点、验收、进度 |
| 2 | [PRD.md](./PRD.md) | §0 规则、§3 契约 |
| 3 | [ARCHITECTURE.md](./ARCHITECTURE.md) | Pod/端口/存储 |
| 4 | [CONVENTIONS.md](./CONVENTIONS.md) | 开发/联调 checklist |
| 5 | [BASTION-DEV.md](./BASTION-DEV.md) | **仅联调阶段** |

## 开发节奏（G-006）

**全部产品需求（30 条 REQ）代码写完 → 再联调。** 见 [dev-then-integrate.mdc](../.cursor/rules/dev-then-integrate.mdc)。

## 禁止

- 开发阶段改 Dockerfile / `buildall.sh` / `helm/` 或上跳板机
- 只做阶段 A 就联调（除非用户明确缩小范围）
- executor `import pkg/plugins/<name>`；`switch plan.Plugin`
- 海量日志/指标写入 MySQL
- 联调通过却不更新 DEVELOPMENT_PLAN

## 环境

| 用途 | 位置 |
|------|------|
| 写代码 | 本机仓库 |
| 联调 | 跳板机 `/root/max/aiops` → [BASTION-DEV.md](./BASTION-DEV.md) |

README 愿景（HelpDesk、巡检、IM 等）对应 **阶段 C/D REQ**，见 DEVELOPMENT_PLAN。
