# AIOps Project Memory

## Project Overview
- AIOps: LLM-powered intelligent operations platform
- Backend: Go (Gin + GORM + MySQL), deployed on K8s
- Frontend: React 18 + TypeScript + Vite 5 + Ant Design 5
- 30 backend REQs + 28 frontend REQs (Phase A-D), all code completed

## Architecture
- Microservices: gateway/platform/chat/policy/executor + plugins(mock/k8s/prometheus/logs) + module-kb + worker
- Key pattern: gateway JWT → service headers (X-AIOps-User-Id, X-Request-Id, etc.)
- Policy engine: boundary.Evaluate → ALLOW/ASK/DENY
- Executor is sole writer of execution_records, must re-evaluate before execution

## Phase E Implementation (2026-05-27)
- PRD optimization: added 9 guard rules (G-007~G-015), 14 new REQs (090-103)
- Backend: 14 REQs implemented (17 new files + 12 modified files), go build/vet/test pass
- Frontend: 12 REQs (F-040~F-051) implemented (16 new files + 8 modified), tsc + build pass
- QA Round 1: found 24 API path mismatches, 5 type inconsistencies, 2 backend bugs
- Fix round: backend added 10 endpoints + fixed SLA severity bug + crypto mutex; frontend aligned all paths + types
- QA Round 2 (self-verified): all pass, found 1 remaining audit path mismatch (fixed)
- Final: go build/vet/test ✅, tsc --noEmit ✅, npm run build ✅

## Key Guard Rules
- G-007: Testable acceptance criteria
- G-008: Unified error format {code, message, request_id}
- G-009: Data validation on all model fields
- G-010: REST naming + idempotent 409
- G-011: 30s timeout, retry ≤3 + backoff, 4xx no retry
- G-012: Structured JSON logging (zap)
- G-013: DB writes use transactions, concurrency uses sync.Map
- G-014: Test coverage requirements
- G-015: No time.Sleep/hardcoded IPs/key leaks
