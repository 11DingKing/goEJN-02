# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

抢修队重复接单时接口报 500，请帮我修复。

工单派发后被 repair-li 接单，另一个抢修队再调 POST /api/v1/workorders/{id}/accept，接口返回 500 Internal Server Error，响应体是 {"error":"work order already accepted: work order wo-1 already accepted by repair-li"}。这在业务上是一次冲突，不是服务器故障，前端按 5xx 走了"系统异常，请稍后重试"的兜底提示，抢修队就一直重试。我们上层还用 errors.Is 判断这类冲突来决定要不要提示"该工单已被他人接单"，现在也判不出来。

对比：同一个工单在只派发未接单时直接调 complete，接口正确返回 409；原接单人用同一个 assignee 重复接单仍然是 200 幂等，这两处是对的。

期望：他人已接单的重复接单要返回 409，并且这个错误能被 errors.Is 识别为对应的业务哨兵错误。修复后请保证 go test ./... 全绿，不要修改或跳过测试。

## 含 Bug 版本

- 仓库：11DingKing/goEJN-02
- 仓库地址：https://github.com/11DingKing/goEJN-02.git
- parent SHA：d641576a5ca7f790cccaa405fce53ae789a94f09

## 复现步骤

```bash
git clone -- https://github.com/11DingKing/goEJN-02.git bug-repro
cd bug-repro
git checkout --detach d641576a5ca7f790cccaa405fce53ae789a94f09
go test ./... -run "^TestDuplicateAccept" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./... -run "^TestDuplicateAccept" -count=1 -v
?   	ejina-microgrid/cmd/server	[no test files]
=== RUN   TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted
    duplicate_accept_test.go:31: errors.Is(err, domain.ErrAlreadyAccepted) = false, err = work order already accepted: work order wo-1 already accepted by repair-li
--- FAIL: TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted (0.00s)
FAIL
FAIL	ejina-microgrid/internal/app	0.044s
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/config	0.046s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/domain	0.042s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/scheduler	0.052s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/store	0.051s [no tests to run]
=== RUN   TestDuplicateAcceptReturnsConflict
    duplicate_accept_test.go:38: second accept by another team: status = 500, want 409; body = {"error":"work order already accepted: work order wo-1 already accepted by repair-li"}
--- FAIL: TestDuplicateAcceptReturnsConflict (0.01s)
FAIL
FAIL	ejina-microgrid/internal/transport/http	0.068s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./... -run "^TestDuplicateAccept" -count=1 -v
?   	ejina-microgrid/cmd/server	[no test files]
=== RUN   TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted
    duplicate_accept_test.go:31: errors.Is(err, domain.ErrAlreadyAccepted) = false, err = work order already accepted: work order wo-1 already accepted by repair-li
--- FAIL: TestDuplicateAcceptByAnotherTeamIsAlreadyAccepted (0.00s)
FAIL
FAIL	ejina-microgrid/internal/app	0.002s
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/config	0.002s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/domain	0.002s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/scheduler	0.002s [no tests to run]
testing: warning: no tests to run
PASS
ok  	ejina-microgrid/internal/store	0.003s [no tests to run]
=== RUN   TestDuplicateAcceptReturnsConflict
    duplicate_accept_test.go:38: second accept by another team: status = 500, want 409; body = {"error":"work order already accepted: work order wo-1 already accepted by repair-li"}
--- FAIL: TestDuplicateAcceptReturnsConflict (0.00s)
FAIL
FAIL	ejina-microgrid/internal/transport/http	0.002s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

定向测试通过：go test ./... -run '^TestDuplicateAccept' -count=1 -v（同时覆盖 service 层错误链与 HTTP 409 两处断言）
全量回归 go test -timeout=120s -count=1 ./... 通过，go build ./... 与 go vet ./... 通过
他人重复接单时 errors.Is(err, domain.ErrAlreadyAccepted) 为真且 HTTP 返回 409；原接单人重复接单仍为 200 幂等；既有 ErrInvalidTransition→409 映射不变；不得修改或跳过既有测试
