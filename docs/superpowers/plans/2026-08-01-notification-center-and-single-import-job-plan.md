# 消息中心与单文件导入后台任务化实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增前端消息中心，并将单文件导入改为后台任务，导入完成后通过消息中心提示用户而非强制跳转。

**Architecture:** 后端 `JobService` 在任务终态发送 `notification:new` 事件；前端新增独立的 `notifications` Pinia store 和消息中心 UI 组件；单文件导入新增 `single_import` 任务类型和 handler，复用现有 `FileService.ImportFile` 完成实际导入。

**Tech Stack:** Go (Wails v2), Vue 3 + TypeScript + Pinia, SQLite, Wails runtime events.

## Global Constraints

- 通知当前阶段仅存于前端内存（A+ 方案），应用关闭后清空。
- 消息中心与任务中心解耦：任务中心持久化任务历史，消息中心负责当前会话即时提醒。
- 通知事件 ID 由后端生成，格式预留后端持久化演进空间。
- 单文件导入与批量导入共用 `ExclusiveKey: "import"`，避免同时执行。
- 单文件导入任务不可恢复（`Recoverable: false`），失败/取消即终态。
- 每个 task 完成后必须跑对应测试，通过后再 commit。
- 所有 Go commit message 以 `Co-Authored-By: Claude <noreply@anthropic.com>` 结尾。

---

## 文件清单

| 文件 | 动作 | 职责 |
|------|------|------|
| `app/models/notification.go` | 创建 | 通知事件类型定义 |
| `app/models/job.go` | 修改 | 新增 `JobTypeSingleImport` 常量 |
| `app/services/single_import_handler.go` | 创建 | 单文件导入任务处理器 |
| `app/services/job_service.go` | 修改 | 注册策略/handler、发送通知事件、新增提交方法 |
| `app/services/job_service_test.go` | 修改 | 通知事件测试 |
| `app/app.go` | 修改 | 新增 `SubmitSingleImportJob` Wails 绑定 |
| `frontend/src/types/notifications.ts` | 创建 | 前端通知类型 |
| `frontend/src/types/jobs.ts` | 修改 | 新增 `SingleImportJobRequest` |
| `frontend/src/store/modules/notifications.ts` | 创建 | 通知 store |
| `frontend/src/components/NotificationBell.vue` | 创建 | 消息中心入口铃铛 |
| `frontend/src/components/NotificationPanel.vue` | 创建 | 下拉消息面板 |
| `frontend/src/components/AppHeader.vue` | 修改 | 集成 NotificationBell |
| `frontend/src/main.ts` | 修改 | 初始化 notifications store |
| `frontend/src/api/index.ts` | 修改 | 新增 `submitSingleImportJob` 绑定 |
| `frontend/src/views/ImportView.vue` | 修改 | 单文件导入改为提交后台任务，移除强制跳转 |

---

### Task 1: 后端通知事件模型与 JobService 通知发送

**Files:**
- Create: `app/models/notification.go`
- Modify: `app/services/job_service.go`
- Test: `app/services/job_service_test.go`

**Interfaces:**
- Consumes: existing `JobService.emitter`, `models.Job`
- Produces: `models.NotificationEvent` struct, `JobService.emitNotificationEvent(job *models.Job)`

- [ ] **Step 1: 创建通知事件模型**

Create `app/models/notification.go`:

```go
package models

type NotificationType string

const (
	NotificationTypeJobCompleted NotificationType = "job_completed"
	NotificationTypeJobFailed    NotificationType = "job_failed"
	NotificationTypeJobCancelled NotificationType = "job_cancelled"
)

type NotificationEvent struct {
	ID        string                 `json:"id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt string                 `json:"createdAt"`
	Read      bool                   `json:"read"`
}
```

- [ ] **Step 2: 在 JobService 新增通知事件辅助方法**

在 `app/services/job_service.go` 中，新增以下私有方法（放在 `emitJobEvent` 附近）：

```go
func (s *JobService) emitNotificationEvent(job *models.Job) {
	if s.emitter == nil {
		return
	}

	notification := models.NotificationEvent{
		ID:        fmt.Sprintf("notif-%d-%d", job.ID, time.Now().UnixMilli()),
		Type:      s.notificationTypeForJob(job),
		Title:     s.notificationTitleForJob(job),
		Message:   s.notificationMessageForJob(job),
		Payload:   map[string]interface{}{"jobId": job.ID, "jobType": job.JobType},
		CreatedAt: time.Now().Format(time.RFC3339),
		Read:      false,
	}

	s.emitter("notification:new", notification)
}

func (s *JobService) notificationTypeForJob(job *models.Job) models.NotificationType {
	switch job.Status {
	case models.JobStatusCompleted:
		return models.NotificationTypeJobCompleted
	case models.JobStatusFailed:
		return models.NotificationTypeJobFailed
	case models.JobStatusCancelled:
		return models.NotificationTypeJobCancelled
	default:
		return models.NotificationTypeJobCompleted
	}
}

func (s *JobService) notificationTitleForJob(job *models.Job) string {
	isSingle := job.JobType == models.JobTypeSingleImport
	switch job.Status {
	case models.JobStatusCompleted:
		if isSingle {
			return "文件导入成功"
		}
		return "批量导入完成"
	case models.JobStatusFailed:
		if isSingle {
			return "文件导入失败"
		}
		return "批量导入失败"
	case models.JobStatusCancelled:
		if isSingle {
			return "文件导入已取消"
		}
		return "批量导入已取消"
	default:
		return "任务状态更新"
	}
}

func (s *JobService) notificationMessageForJob(job *models.Job) string {
	if job.Status == models.JobStatusCompleted && job.JobType == models.JobTypeSingleImport {
		if job.Title != "" {
			return job.Title
		}
		return "文件导入成功"
	}
	if job.Status == models.JobStatusCompleted && job.JobType == models.JobTypeBatchImport {
		var result models.BatchImportResult
		if err := json.Unmarshal(job.Result, &result); err == nil {
			return fmt.Sprintf("成功导入 %d 个文件，失败 %d 个", result.SuccessCount, result.FailedCount)
		}
		return "批量导入已完成"
	}
	if job.ErrorMessage != "" {
		return job.ErrorMessage
	}
	return job.ProgressMessage
}
```

- [ ] **Step 3: 在任务终态处调用通知发送**

在 `app/services/job_service.go` 的 `runJob` 方法中：

找到成功完成后的代码块：

```go
job.Status = models.JobStatusCompleted
job.ErrorMessage = ""
job.ProgressCompleted = job.ProgressTotal
job.FinishedAt = time.Now().Format(time.RFC3339)
if err := s.jobRepo.Update(job); err != nil {
    return
}
s.emitJobEvent("job:completed", job)
```

在其后插入：

```go
s.emitNotificationEvent(job)
```

找到失败处理后的代码块：

```go
job.Status = models.JobStatusFailed
job.ErrorMessage = err.Error()
job.FinishedAt = time.Now().Format(time.RFC3339)
_ = s.jobRepo.Update(job)
s.emitJobEvent("job:failed", job)
return
```

在其后插入：

```go
s.emitNotificationEvent(job)
```

找到取消处理分支（`context.Canceled` 中 `job.Status == models.JobStatusCancelled`）：

```go
case models.JobStatusCancelled, models.JobStatusAwaitingResume, models.JobStatusTimedOut:
    s.emitJobEvent("job:updated", job)
    return
```

在 `return` 前，对 `cancelled` 单独插入：

```go
if job.Status == models.JobStatusCancelled {
    s.emitNotificationEvent(job)
}
```

> 说明：`timed_out` 由 `watchTimeouts` 设置并触发 `job:failed`，已在前面的失败通知逻辑中覆盖。
>

- [ ] **Step 4: 编写测试验证通知事件**

在 `app/services/job_service_test.go` 新增：

```go
type successTestHandler struct{}

func (h *successTestHandler) Type() string { return "success_test" }
func (h *successTestHandler) Validate(payload json.RawMessage) error { return nil }
func (h *successTestHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}
func (h *successTestHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}
func (h *successTestHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return nil
}

func TestJobService_EmitsNotificationOnCompletion(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	service := NewJobService(jobRepo, nil, nil, nil)
	service.RegisterHandler(&successTestHandler{})
	service.policies["success_test"] = JobPolicy{
		JobType:          "success_test",
		MaxConcurrent:    1,
		CanRunBackground: true,
		Recoverable:      false,
		Timeout:          time.Minute,
	}

	var emitted []models.NotificationEvent
	service.SetEventEmitter(func(eventName string, data interface{}) {
		if eventName == "notification:new" {
			if n, ok := data.(models.NotificationEvent); ok {
				emitted = append(emitted, n)
			}
		}
	})

	job := &models.Job{
		JobType:       "success_test",
		Status:        models.JobStatusPending,
		Title:         "test",
		Payload:       json.RawMessage("{}"),
		ProgressTotal: 1,
		CanResume:     false,
	}
	if err := jobRepo.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	service.runJob(job.ID, false)

	if len(emitted) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(emitted))
	}
	if emitted[0].Type != models.NotificationTypeJobCompleted {
		t.Errorf("expected type completed, got %s", emitted[0].Type)
	}
	if emitted[0].ID == "" {
		t.Error("expected notification id to be set")
	}
}
```

- [ ] **Step 5: 运行测试**

Run: `go test ./app/services -run TestJobService_EmitsNotification -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add app/models/notification.go app/services/job_service.go app/services/job_service_test.go
git commit -m "feat(backend): emit notification events on job completion and failure

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 2: 单文件导入 Handler

**Files:**
- Create: `app/services/single_import_handler.go`
- Modify: `app/models/job.go`
- Test: `app/services/single_import_handler_test.go`

**Interfaces:**
- Consumes: `services.ImportFileRequest`, `JobRuntime.FileService()`
- Produces: `SingleImportHandler` implementing `JobHandler`

- [ ] **Step 1: 新增单文件导入任务类型常量**

在 `app/models/job.go` 中：

```go
const (
	JobTypeBatchImport  = "batch_import"
	JobTypeSingleImport = "single_import"
)
```

- [ ] **Step 2: 创建 SingleImportHandler**

Create `app/services/single_import_handler.go`:

```go
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"LocalSpace/app/models"
)

type SingleImportHandler struct{}

func NewSingleImportHandler() *SingleImportHandler {
	return &SingleImportHandler{}
}

func (h *SingleImportHandler) Type() string {
	return models.JobTypeSingleImport
}

func (h *SingleImportHandler) Validate(payload json.RawMessage) error {
	var req ImportFileRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid single import payload: %w", err)
	}
	if req.FilePath == "" {
		return fmt.Errorf("file path is required")
	}
	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}
	return nil
}

func (h *SingleImportHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	var req ImportFileRequest
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal single import payload: %w", err)
	}

	fileService := runtime.FileService()
	if fileService == nil {
		return fmt.Errorf("file service is not available")
	}

	_ = runtime.UpdateHeartbeat("正在导入文件")
	if err := fileService.ImportFile(req); err != nil {
		return err
	}

	_ = runtime.UpdateHeartbeat("导入完成")
	return nil
}

func (h *SingleImportHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	return fmt.Errorf("single import does not support resume")
}

func (h *SingleImportHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
	// SingleImportHandler uses FileService.ImportFile which already handles staging cleanup.
	return nil
}
```

- [ ] **Step 3: 编写 handler 测试**

Create `app/services/single_import_handler_test.go`:

```go
package services

import (
	"context"
	"encoding/json"
	"testing"

	"LocalSpace/app/models"
)

func TestSingleImportHandler_Type(t *testing.T) {
	h := NewSingleImportHandler()
	if h.Type() != models.JobTypeSingleImport {
		t.Errorf("expected type %s, got %s", models.JobTypeSingleImport, h.Type())
	}
}

func TestSingleImportHandler_Validate(t *testing.T) {
	h := NewSingleImportHandler()

	err := h.Validate(json.RawMessage(`{"file_path": "/tmp/a.txt", "file_name": "a.txt"}`))
	if err != nil {
		t.Errorf("expected valid payload, got %v", err)
	}

	err = h.Validate(json.RawMessage(`{"file_path": "", "file_name": ""}`))
	if err == nil {
		t.Error("expected validation error for empty fields")
	}
}

func TestSingleImportHandler_Resume_ReturnsError(t *testing.T) {
	h := NewSingleImportHandler()
	err := h.Resume(context.Background(), &models.Job{}, nil)
	if err == nil {
		t.Error("expected resume to return error")
	}
}
```

- [ ] **Step 4: 运行测试**

Run: `go test ./app/services -run TestSingleImportHandler -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/models/job.go app/services/single_import_handler.go app/services/single_import_handler_test.go
git commit -m "feat(backend): add single import job handler

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 3: 注册单文件导入策略与 Wails 绑定

**Files:**
- Modify: `app/services/job_service.go`
- Modify: `app/app.go`
- Test: `app/services/job_service_test.go`

**Interfaces:**
- Consumes: `SingleImportHandler`, `models.JobTypeSingleImport`
- Produces: `JobService.SubmitSingleImportJob(payload json.RawMessage) (*models.Job, error)`, `App.SubmitSingleImportJob(...)`

- [ ] **Step 1: 在 JobService 中注册策略和 handler**

在 `app/services/job_service.go` 的 `NewJobService` 中，修改 `policies`：

```go
policies: map[string]JobPolicy{
    models.JobTypeBatchImport: {
        JobType:          models.JobTypeBatchImport,
        MaxConcurrent:    1,
        ExclusiveKey:     "import",
        CanRunBackground: true,
        Recoverable:      true,
        Timeout:          2 * time.Hour,
    },
    models.JobTypeSingleImport: {
        JobType:          models.JobTypeSingleImport,
        MaxConcurrent:    1,
        ExclusiveKey:     "import",
        CanRunBackground: true,
        Recoverable:      false,
        Timeout:          30 * time.Minute,
    },
},
```

并注册 handler：

```go
service.RegisterHandler(NewBatchImportHandler())
service.RegisterHandler(NewSingleImportHandler())
```

- [ ] **Step 2: 新增 SubmitSingleImportJob 方法**

在 `app/services/job_service.go` 中，放在 `SubmitBatchImportJob` 之后：

```go
func (s *JobService) SubmitSingleImportJob(req ImportFileRequest) (*models.Job, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal single import payload: %w", err)
	}

	handler, policy, err := s.handlerAndPolicy(models.JobTypeSingleImport)
	if err != nil {
		return nil, err
	}
	if err := handler.Validate(payload); err != nil {
		return nil, err
	}

	if err := s.ensureConcurrency(policy, 0); err != nil {
		return nil, err
	}

	job := &models.Job{
		JobType:           models.JobTypeSingleImport,
		Status:            models.JobStatusPending,
		Title:             fmt.Sprintf("导入 %s", req.FileName),
		Payload:           payload,
		ProgressTotal:     1,
		ProgressCompleted: 0,
		ProgressMessage:   "等待开始导入",
		ExclusiveKey:      policy.ExclusiveKey,
		CanResume:         policy.Recoverable,
	}
	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	s.emitJobEvent("job:created", job)
	go s.runJob(job.ID, false)

	return job, nil
}
```

- [ ] **Step 3: 在 App 中暴露 Wails 绑定**

在 `app/app.go` 中，放在 `ImportFileWithMetadata` 之后（约第 280 行）：新增：

```go
// SubmitSingleImportJob creates a background job for importing a single file.
func (a *App) SubmitSingleImportJob(
	filePath string,
	fileName string,
	description string,
	tags []string,
	keywords string,
	collectionName string,
) (*models.Job, error) {
	return a.jobService.SubmitSingleImportJob(services.ImportFileRequest{
		FilePath:       filePath,
		FileName:       fileName,
		Description:    description,
		Tags:           tags,
		Keywords:       keywords,
		CollectionName: collectionName,
	})
}
```

- [ ] **Step 4: 编写提交测试**

在 `app/services/job_service_test.go` 新增：

```go
func TestJobService_SubmitSingleImportJob_CreatesJob(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	jobRepo := repositories.NewJobRepository(repositories.NewSQLiteDBWrapper(db))
	jobService := NewJobService(jobRepo, nil, nil, nil)

	var createdJob *models.Job
	jobService.SetEventEmitter(func(eventName string, data interface{}) {
		if eventName == "job:created" {
			if j, ok := data.(*models.Job); ok {
				createdJob = j
			}
		}
	})

	req := ImportFileRequest{
		FilePath: "/tmp/test.txt",
		FileName: "test.txt",
		Tags:     []string{"test"},
	}
	job, err := jobService.SubmitSingleImportJob(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if job.JobType != models.JobTypeSingleImport {
		t.Errorf("expected job type %s, got %s", models.JobTypeSingleImport, job.JobType)
	}
	if createdJob == nil {
		t.Error("expected job:created event to be emitted")
	}
}
```

- [ ] **Step 5: 运行测试**

Run: `go test ./app/services -run TestJobService_SubmitSingleImportJob -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add app/services/job_service.go app/app.go app/services/job_service_test.go
git commit -m "feat(backend): register single import policy and expose Wails binding

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 4: 前端通知类型与 Store

**Files:**
- Create: `frontend/src/types/notifications.ts`
- Create: `frontend/src/store/modules/notifications.ts`

**Interfaces:**
- Consumes: Wails `notification:new` event
- Produces: `useNotificationsStore` with `items`, `unreadCount`, `addNotification`, `markAsRead`, `markAllAsRead`, `removeNotification`, `clearAll`, `initialize`

- [ ] **Step 1: 创建通知类型文件**

Create `frontend/src/types/notifications.ts`:

```ts
export interface NotificationEvent {
  id: string
  type: 'job_completed' | 'job_failed' | 'job_cancelled'
  title: string
  message: string
  payload: {
    jobId?: number
    jobType?: string
    [key: string]: any
  }
  createdAt: string
  read: boolean
}
```

- [ ] **Step 2: 创建 notifications store**

Create `frontend/src/store/modules/notifications.ts`:

```ts
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { isWailsAvailable } from '@/api'
import type { NotificationEvent } from '@/types/notifications'

const MAX_NOTIFICATIONS = 50

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationEvent[]>([])

  const unreadCount = computed(() => items.value.filter((item) => !item.read).length)
  const hasUnread = computed(() => unreadCount.value > 0)

  const addNotification = (notification: NotificationEvent) => {
    items.value.unshift(notification)
    if (items.value.length > MAX_NOTIFICATIONS) {
      items.value.length = MAX_NOTIFICATIONS
    }
  }

  const markAsRead = (id: string) => {
    const item = items.value.find((n) => n.id === id)
    if (item) {
      item.read = true
    }
  }

  const markAllAsRead = () => {
    items.value.forEach((item) => {
      item.read = true
    })
  }

  const removeNotification = (id: string) => {
    items.value = items.value.filter((item) => item.id !== id)
  }

  const clearAll = () => {
    items.value = []
  }

  const initialize = () => {
    if (!isWailsAvailable()) {
      return
    }

    EventsOn('notification:new', (notification: NotificationEvent) => {
      addNotification(notification)
    })
  }

  return {
    items,
    unreadCount,
    hasUnread,
    addNotification,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAll,
    initialize,
  }
})
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/notifications.ts frontend/src/store/modules/notifications.ts
git commit -m "feat(frontend): add notifications store and types

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 5: 消息中心 UI 组件与 AppHeader 集成

**Files:**
- Create: `frontend/src/components/NotificationBell.vue`
- Create: `frontend/src/components/NotificationPanel.vue`
- Modify: `frontend/src/components/AppHeader.vue`
- Modify: `frontend/src/main.ts`

**Interfaces:**
- Consumes: `useNotificationsStore`
- Produces: `NotificationBell`, `NotificationPanel` components

- [ ] **Step 1: 创建 NotificationPanel.vue**

Create `frontend/src/components/NotificationPanel.vue`:

```vue
<template>
  <div v-if="visible" class="notification-panel">
    <div class="notification-panel-header">
      <strong>消息中心</strong>
      <div class="notification-actions">
        <button type="button" class="text-btn" @click="markAllAsRead">全部已读</button>
        <button type="button" class="text-btn" @click="clearAll">清空</button>
      </div>
    </div>

    <div v-if="items.length === 0" class="notification-empty">暂无消息</div>

    <ul v-else class="notification-list">
      <li
        v-for="item in items"
        :key="item.id"
        :class="['notification-item', { unread: !item.read }]"
        @click="handleClick(item)"
      >
        <div class="notification-dot" v-if="!item.read"></div>
        <div class="notification-content">
          <div class="notification-title">{{ item.title }}</div>
          <div class="notification-message">{{ item.message }}</div>
          <div class="notification-time">{{ formatTime(item.createdAt) }}</div>
        </div>
        <button
          type="button"
          class="notification-close"
          @click.stop="removeNotification(item.id)"
        >
          ×
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useNotificationsStore } from '@/store/modules/notifications'
import type { NotificationEvent } from '@/types/notifications'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'navigate', path: string): void
}>()

const { items, markAllAsRead, markAsRead, removeNotification, clearAll } = useNotificationsStore()

const handleClick = (item: NotificationEvent) => {
  markAsRead(item.id)
  if (item.type === 'job_completed' && item.payload?.jobId) {
    emit('navigate', '/tasks')
  }
}

const formatTime = (iso: string) => {
  const date = new Date(iso)
  return date.toLocaleString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<style scoped>
.notification-panel {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 360px;
  max-height: 420px;
  background: var(--surface-color, #fff);
  border: 1px solid var(--border-color, rgba(148, 163, 184, 0.16));
  border-radius: 18px;
  box-shadow: 0 24px 48px rgba(44, 62, 94, 0.12);
  display: flex;
  flex-direction: column;
  z-index: 1000;
  overflow: hidden;
}

.notification-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color, rgba(148, 163, 184, 0.16));
  font-size: 14px;
}

.notification-actions {
  display: flex;
  gap: 12px;
}

.text-btn {
  background: none;
  border: none;
  color: var(--primary-color, #2196f3);
  font-size: 12px;
  cursor: pointer;
}

.notification-empty {
  padding: 32px 16px;
  text-align: center;
  color: var(--text-faint, #94a3b8);
  font-size: 13px;
}

.notification-list {
  list-style: none;
  margin: 0;
  padding: 8px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border-radius: 14px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.notification-item:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.08));
}

.notification-item.unread {
  background: var(--unread-bg, rgba(33, 150, 243, 0.04));
}

.notification-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary-color, #2196f3);
  margin-top: 6px;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-color, #1e293b);
}

.notification-message {
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-soft, #64748b);
  line-height: 1.5;
  word-break: break-word;
}

.notification-time {
  margin-top: 6px;
  font-size: 11px;
  color: var(--text-faint, #94a3b8);
}

.notification-close {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  background: none;
  color: var(--text-faint, #94a3b8);
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  flex-shrink: 0;
}

.notification-close:hover {
  background: var(--hover-bg, rgba(148, 163, 184, 0.12));
}
</style>
```

- [ ] **Step 2: 创建 NotificationBell.vue**

Create `frontend/src/components/NotificationBell.vue`:

```vue
<template>
  <div class="notification-bell-wrapper" ref="wrapperRef">
    <button
      type="button"
      class="notification-bell"
      :class="{ active: panelVisible }"
      aria-label="消息中心"
      @click="togglePanel"
    >
      <span class="bell-icon">🔔</span>
      <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>
    <NotificationPanel
      :visible="panelVisible"
      @navigate="handleNavigate"
      @click.stop
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationsStore } from '@/store/modules/notifications'
import NotificationPanel from './NotificationPanel.vue'

const { unreadCount } = useNotificationsStore()
const router = useRouter()

const panelVisible = ref(false)
const wrapperRef = ref<HTMLElement | null>(null)

const togglePanel = () => {
  panelVisible.value = !panelVisible.value
}

const handleNavigate = (path: string) => {
  panelVisible.value = false
  router.push(path)
}

const handleClickOutside = (event: MouseEvent) => {
  if (wrapperRef.value && !wrapperRef.value.contains(event.target as Node)) {
    panelVisible.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.notification-bell-wrapper {
  position: relative;
}

.notification-bell {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s ease;
}

.notification-bell:hover,
.notification-bell.active {
  background: var(--hover-bg, rgba(148, 163, 184, 0.12));
}

.bell-icon {
  font-size: 16px;
}

.unread-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>
```

- [ ] **Step 3: 在 AppHeader.vue 集成铃铛**

修改 `frontend/src/components/AppHeader.vue`：

在 `<script setup>` 中 import：

```ts
import NotificationBell from '@/components/NotificationBell.vue'
```

修改 `header-actions` div：

```vue
<div class="header-actions">
  <TaskStatusIndicator />
  <NotificationBell />
  <router-link to="/import" :class="['btn', isImportActive ? 'primary' : 'secondary']">
    导入
  </router-link>
  <router-link to="/settings" :class="['btn', isSettingsActive ? 'primary' : 'secondary']">
    设置
  </router-link>
</div>
```

- [ ] **Step 4: 在 main.ts 初始化 notifications store**

修改 `frontend/src/main.ts`：

```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useJobsStore } from './store/modules/jobs'
import { useNotificationsStore } from './store/modules/notifications'
import { useThemeStore } from './store/modules/theme'
import './assets/styles/main.css'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

app.mount('#app')

// Initialize theme after app is mounted
const themeStore = useThemeStore()
themeStore.loadThemeFromBackend()

const jobsStore = useJobsStore()
jobsStore.initialize()

const notificationsStore = useNotificationsStore()
notificationsStore.initialize()
```

- [ ] **Step 5: 运行前端构建检查**

Run: `cd frontend && npm run build`
Expected: 构建成功，无 TypeScript 错误

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/NotificationBell.vue frontend/src/components/NotificationPanel.vue frontend/src/components/AppHeader.vue frontend/src/main.ts
git commit -m "feat(frontend): add notification center UI and integrate into header

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 6: ImportView 改为提交后台任务

**Files:**
- Modify: `frontend/src/views/ImportView.vue`
- Modify: `frontend/src/api/index.ts`
- Modify: `frontend/src/types/jobs.ts`

**Interfaces:**
- Consumes: `api.jobs.submitSingleImportJob`
- Produces: `SingleImportJobRequest` type, non-blocking single import submission

- [ ] **Step 1: 新增 SingleImportJobRequest 类型**

在 `frontend/src/types/jobs.ts` 中新增：

```ts
export interface SingleImportJobRequest {
  filePath: string
  fileName: string
  description: string
  tags: string[]
  keywords: string
  collectionName: string
}
```

- [ ] **Step 2: 在 api/index.ts 新增绑定**

在 `frontend/src/api/index.ts` 的 `jobs` 对象中新增：

```ts
submitSingleImportJob: (payload: SingleImportJobRequest) =>
  new Promise((resolve, reject) => {
    try {
      if (!window.go || !window.go.app || !window.go.app.App) {
        reject(new Error('Wails API not available'))
        return
      }
      window.go!.app!.App.SubmitSingleImportJob(
        payload.filePath,
        payload.fileName,
        payload.description,
        payload.tags,
        payload.keywords,
        payload.collectionName
      )
        .then(resolve)
        .catch(reject)
    } catch (error) {
      reject(error)
    }
  }),
```

并在文件顶部 import `SingleImportJobRequest`：

```ts
import type { BatchImportJobRequest, SingleImportJobRequest } from '@/types/jobs'
```

- [ ] **Step 3: 修改 ImportView.vue 的 submitSingleImport**

在 `frontend/src/views/ImportView.vue` 中：

修改 import 的类型：

```ts
import type { BatchImportJobRequest, SelectedFile, SingleImportJobRequest } from '@/types/jobs'
```

替换 `submitSingleImport` 函数为：

```ts
const submitSingleImport = async () => {
  if (!singleFile.value || !singleCanSubmit.value) return

  singleSubmitting.value = true
  singleError.value = ''
  singleSuccess.value = ''

  try {
    const payload: SingleImportJobRequest = {
      filePath: singleFile.value.path,
      fileName: singleForm.fileName.trim(),
      description: singleForm.description.trim(),
      tags: singleForm.tags,
      keywords: singleForm.keywords.trim(),
      collectionName: singleForm.collectionName.trim(),
    }

    await api.jobs.submitSingleImportJob(payload)
    singleSuccess.value = '已创建后台导入任务，可在右上角消息中心或任务中心查看进度。'
    singleFile.value = null
    resetSingleForm()
  } catch (error) {
    singleError.value = error instanceof Error ? error.message : '提交导入任务失败'
  } finally {
    singleSubmitting.value = false
  }
}
```

确保移除了原函数中的 `setTimeout(() => { router.push('/files') }, 400)`。

同时检查 `ImportView.vue` 是否仍需要 `useRouter`，如果没有其他地方使用，可以移除 import。如果批量任务提交后仍使用 `router`，则保留。

- [ ] **Step 4: 运行前端构建检查**

Run: `cd frontend && npm run build`
Expected: 构建成功

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/ImportView.vue frontend/src/api/index.ts frontend/src/types/jobs.ts
git commit -m "feat(frontend): convert single file import to background job submission

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 7: 集成验证

**Files:**
- Modify: 无新文件，纯验证

**Interfaces:**
- 验证后端测试、前端构建、Wails 绑定生成

- [ ] **Step 1: 运行所有后端服务测试**

Run: `go test ./app/services ./app/repositories ./app/app -v`
Expected: PASS

- [ ] **Step 2: 检查 Wails 绑定是否需要重新生成**

如果项目使用 `wailsjs` 自动绑定生成：

Run: `wails generate module` 或 `wails dev`（视项目配置而定）
Expected: 生成 `frontend/wailsjs/go/app/App.ts` 包含 `SubmitSingleImportJob` 方法

> 如果 `frontend/wailsjs` 是自动生成的绑定目录，确认 `SubmitSingleImportJob` 已出现在 `frontend/wailsjs/go/app/App.ts` 中。

- [ ] **Step 3: 运行前端构建**

Run: `cd frontend && npm run build`
Expected: 构建成功

- [ ] **Step 4: 全量 Go 测试（可选，若时间允许）**

Run: `go test ./...`
Expected: 与上次基线一致，新增测试通过

- [ ] **Step 5: Commit（若绑定文件有更新）**

```bash
git add frontend/wailsjs/
git commit -m "chore: regenerate wails bindings for SubmitSingleImportJob

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## 自我审查清单

1. **Spec coverage:**
   - 消息中心 store + UI → Task 4, Task 5
   - 后端通知事件 → Task 1
   - 单文件导入后台任务化 → Task 2, Task 3, Task 6
   - 不强制跳转 → Task 6
   - 演进预留（后端生成通知 ID） → Task 1

2. **Placeholder scan:** 无 TBD/TODO/"implement later" 等占位。

3. **Type consistency:**
   - `SingleImportJobRequest` 前后端字段一致；
   - `NotificationEvent` 前后端字段一致；
   - `JobTypeSingleImport` 在 Go 中使用一致。

---

## 执行方式选择

Plan complete and saved to `docs/superpowers/plans/2026-08-01-notification-center-and-single-import-job-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
