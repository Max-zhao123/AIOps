# 编码与流程（AI）

PRD §0（**G-006**）：**[DEVELOPMENT_PLAN](./DEVELOPMENT_PLAN.md) 全部 30 条 REQ 代码写完 → 再联调。**

## 开发 vs 联调

| | 开发（30 条齐） | 联调 |
|--|----------------|------|
| 范围 | A+B+C+D 全部 Go 代码 | Dockerfile、Helm、跳板机 |
| 进度 | 更新 PLAN `代码` 列 | 按 PLAN `验收` 列标 `done` |
| 禁止 | 上集群、改 Helm | 为试 Pod 改业务契约 |

## 模块化（G-002）

- 新插件：`pkg/plugins/<n>` + `cmd/plugin-<n>`
- 新功能模块：`pkg/module/<n>` + `cmd/module-<n>` + `aiops-module-<n>`
- executor 禁止 `import` 具体插件

## 开发阶段 checklist

- [ ] 当前 REQ 实现要点（见 PLAN 表）已满足，非 stub
- [ ] `go build` 通过
- [ ] PLAN 已更新

## 联调阶段 checklist

- [ ] 30 条代码均已 **完成**
- [ ] 镜像/Pod 名符合 ARCHITECTURE
- [ ] 该 REQ 验收通过 → `done`
