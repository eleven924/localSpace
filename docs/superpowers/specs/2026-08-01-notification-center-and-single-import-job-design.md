# 消息中心与单文件导入后台任务化设计

## 目标

1. 在前端新增一个**独立的消息中心**，用于展示后台任务完成、失败等事件通知；
2. 将**单文件导入**从同步调用改为**后台任务（JobService）**，支持在任务中心查看进度；
3. 导入完成后不再强制跳转页面，而是通过消息中心提示用户，由用户自行决定下一步；
4. 设计方案需预留向后端持久化通知（方案 B）演进的接口空间。

## 当前问题

- `ImportView.vue` 中 `submitSingleImport` 调用 `api.file.importWithMetadata` 是同步 Wails 绑定，导入大文件或启用 AI 元数据生成时会长时间阻塞前端；
- 导入成功后 400ms 强制 `router.push('/files')`，即使用户已切换页面也会被拉回；
- 前端没有统一的通知机制，后台任务完成状态只能通过任务中心查看，无法做到"即时提醒、不打扰当前页面"。

## 设计原则

1. **消息中心 ≠ 任务中心**
   - 任务中心：持久化记录所有后台任务的完整生命周期、进度、错误详情，用于回溯和手动恢复。
   - 消息中心：当前会话的即时通知列表，负责"发生了什么事"的轻量提示，支持一键跳转。
2. **A+ 阶段：内存级通知**
   - 通知仅存于前端 Pinia store，应用关闭后清空。
   - 任务历史由任务中心持久化，消息中心不再重复存储。
3. **事件驱动**
   - 后端通过 Wails `runtime.EventsEmit` 发送 `notification:new` 事件；
   - 前端 store 监听事件并维护通知列表。
4. **预留演进**
   - 通知 ID、事件字段、store 接口在一开始就设计为后端持久化兼容，后续切换到方案 B 时前端改动最小。

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                          后端 (Go)                           │
│  ┌──────────────┐     ┌──────────────┐     ┌───────────┐   │
│  │ SingleImport │     │ BatchImport  │     │ 未来其他   │   │
│  │   Handler    │     │   Handler    │     │ 任务/事件  │   │
│  └──────┬───────┘     └──────┬───────┘     └─────┬─────┘   │
│         │                    │                    │         │
│         └────────────────────┬────────────────────┘         │
│                            ▼                                │
│                    ┌─────────────────┐                      │
│                    │    JobService   │                      │
│                    │  - 任务状态管理  │                      │
│                    │  - 发送任务事件  │                      │
│                    │  - 发送通知事件  │                      │
│                    └────────┬────────┘                      │
│                             │                               │
│              ┌──────────────┼──────────────┐               │
│              ▼              ▼              ▼                │
│        job:updated    job:completed   notification:new      │
│        job:failed     job:needs-resume                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Wails EventsEmit
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                         前端 (Vue 3)                         │
│  ┌──────────────┐     ┌─────────────────┐   ┌────────────┐ │
│  │   jobs store │     │ notifications   │   │ AppHeader  │ │
│  │              │     │     store       │   │  (铃铛入口) │ │
│  └──────────────┘     └────────┬────────┘   └─────┬──────┘ │
│                                │                   │       │
│                                ▼                   ▼       │
│                       ┌─────────────────┐   ┌───────────┐  │
│                       │ NotificationBell │   │ NotificationPanel │  │
│                       │   + 未读红点      │   │ 下拉消息面板      │  │
│                       └─────────────────┘   └───────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 后端设计

### 1. 新增任务类型

在 `app/models/job.go` 中新增：

```go
const JobTypeSingleImport = "single_import"
```

### 2. 新增 `SingleImportHandler`

位置：`app/services/single_import_handler.go`

职责：
- 处理单个文件的导入流程；
- 调用现有 `FileService.ImportFile` 完成实际的复制、元数据生成、入库；
- 通过 `JobRuntime` 更新进度消息。

```go
type SingleImportHandler struct{}

func NewSingleImportHandler() *SingleImportHandler {
    return &SingleImportHandler{}
}

func (h *SingleImportHandler) Type() string {
    return models.JobTypeSingleImport
}

func (h *SingleImportHandler) Validate(payload json.RawMessage) error {
    var req services.ImportFileRequest
    return json.Unmarshal(payload, &req)
}

func (h *SingleImportHandler) Execute(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    var req services.ImportFileRequest
    if err := json.Unmarshal(job.Payload, &req); err != nil {
        return err
    }

    runtime.SetProgressMessage("开始导入文件")
    fileService := runtime.FileService()
    if err := fileService.ImportFile(req); err != nil {
        return err
    }

    runtime.SetProgressMessage("导入完成")
    return nil
}

func (h *SingleImportHandler) Resume(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    // 单文件导入不可恢复，标记失败
    return fmt.Errorf("single import does not support resume")
}

func (h *SingleImportHandler) Cleanup(ctx context.Context, job *models.Job, runtime JobRuntime) error {
    // 单文件导入的中间状态由 FileService 的 staged pipeline 处理，清理逻辑由 ImportFile 自身负责
    return nil
}
```

> **说明**：`JobRuntime` 需要暴露 `FileService()` 和 `SetProgressMessage(msg string)` 方法。若当前 `JobRuntime` 未暴露 `FileService`，则通过 `JobService` 在创建 `SingleImportHandler` 时注入 `*FileService`。

### 3. JobService 中注册单文件导入策略

在 `NewJobService` 的 `policies` 中新增：

```go
models.JobTypeSingleImport: {
    JobType:          models.JobTypeSingleImport,
    MaxConcurrent:    1,
    ExclusiveKey:     "import",        // 与批量导入共用排他键，避免同时执行
    CanRunBackground: true,
    Recoverable:      false,         // 单文件导入不恢复，失败/取消即终态
    Timeout:          30 * time.Minute,
},
```

注册 handler：

```go
service.RegisterHandler(NewSingleImportHandler())
```

### 4. 新增提交单文件导入任务的 Wails 绑定

在 `app/app.go` 中新增方法：

```go
func (a *App) SubmitSingleImportJob(
    filePath string,
    fileName string,
    description string,
    tags []string,
    keywords string,
    collectionName string,
) (*models.Job, error) {
    req := models.SingleImportJobRequest{
        ImportFileRequest: services.ImportFileRequest{
            FilePath:       filePath,
            FileName:       fileName,
            Description:    description,
            Tags:           tags,
            Keywords:       keywords,
            CollectionName: collectionName,
        },
    }
    payload, _ := json.Marshal(req)
    return a.jobService.SubmitSingleImportJob(payload)
}
```

> `models.SingleImportJobRequest` 可与 `services.ImportFileRequest` 结构对齐，避免重复定义。

### 5. JobService 新增通知事件

在 `runJob` 的任务终态处理点增加通知事件发送：

```go
// 任务成功完成
if job.Status == models.JobStatusCompleted {
    s.emitNotificationEvent(job)
}

// 任务失败
if job.Status == models.JobStatusFailed {
    s.emitNotificationEvent(job)
}
```

新增方法：

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
```

### 6. 通知事件模型

在 `app/models` 下新增 `notification.go`：

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

## 前端设计

### 1. 新增 `notifications` store

位置：`frontend/src/store/modules/notifications.ts`

```ts
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { EventsOn } from '../../../wailsjs/runtime/runtime'
import { isWailsAvailable } from '@/api'
import type { NotificationEvent } from '@/types/notifications'

const MAX_NOTIFICATIONS = 50

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationEvent[]>([])
  const unreadCount = computed(() => items.value.filter((n) => !n.read).length)

  const addNotification = (notification: NotificationEvent) => {
    items.value.unshift(notification)
    if (items.value.length > MAX_NOTIFICATIONS) {
      items.value.length = MAX_NOTIFICATIONS
    }
  }

  const markAsRead = (id: string) => {
    const item = items.value.find((n) => n.id === id)
    if (item) item.read = true
  }

  const markAllAsRead = () => {
    items.value.forEach((n) => (n.read = true))
  }

  const removeNotification = (id: string) => {
    items.value = items.value.filter((n) => n.id !== id)
  }

  const clearAll = () => {
    items.value = []
  }

  const initialize = () => {
    if (!isWailsAvailable()) return

    EventsOn('notification:new', (notification: NotificationEvent) => {
      addNotification(notification)
    })
  }

  return {
    items,
    unreadCount,
    addNotification,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAll,
    initialize,
  }
})
```

### 2. 新增通知类型

位置：`frontend/src/types/notifications.ts`

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

### 3. 消息中心 UI

新增两个组件：

#### `NotificationBell.vue`

- 显示铃铛图标；
- 未读数红点；
- 点击切换 `NotificationPanel` 显示/隐藏。

位置：`frontend/src/components/NotificationBell.vue`

#### `NotificationPanel.vue`

- 下拉面板，列出最近通知；
- 每条通知显示标题、消息、时间；
- 支持"标记已读"、"全部已读"、"清空"；
- 支持点击通知触发跳转（如 job_completed 显示"去查看"按钮）。

位置：`frontend/src/components/NotificationPanel.vue`

### 4. AppHeader 集成

在 `AppHeader.vue` 中，在 `TaskStatusIndicator` 旁边插入 `NotificationBell`：

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

### 5. 应用启动时初始化通知 store

在 `frontend/src/main.ts` 中：

```ts
import { useNotificationsStore } from '@/store/modules/notifications'

const app = createApp(App)
// ... 注册 pinia、router

const jobsStore = useJobsStore(pinia)
const notificationsStore = useNotificationsStore(pinia)

jobsStore.initialize()
notificationsStore.initialize()
```

### 6. ImportView 改造

`submitSingleImport` 改为提交后台任务：

```ts
const submitSingleImport = async () => {
  if (!singleFile.value || !singleCanSubmit.value) return

  singleSubmitting.value = true
  singleError.value = ''
  singleSuccess.value = ''

  try {
    await api.jobs.submitSingleImportJob({
      filePath: singleFile.value.path,
      fileName: singleForm.fileName.trim(),
      description: singleForm.description.trim(),
      tags: singleForm.tags,
      keywords: singleForm.keywords.trim(),
      collectionName: singleForm.collectionName.trim(),
    })

    singleSuccess.value = '已创建后台导入任务，可在右上角消息中心或任务中心查看进度。'
    resetSingleForm()
    singleFile.value = null
  } catch (error) {
    singleError.value = error instanceof Error ? error.message : '提交导入任务失败'
  } finally {
    singleSubmitting.value = false
  }
}
```

同时移除原来导入成功后的 `setTimeout(() => router.push('/files'), 400)` 强制跳转。

### 7. api 层新增绑定

在 `frontend/src/api/index.ts` 中新增：

```ts
jobs: {
  // ... 已有方法
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
}
```

新增类型：

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

## 通知内容规则

| 任务类型 | 任务状态 | 通知标题 | 通知消息示例 |
|---------|---------|---------|------------|
| `single_import` | completed | 文件导入成功 | "movie.mp4" 已成功导入 |
| `single_import` | failed | 文件导入失败 | "movie.mp4" 导入失败：无法读取源文件 |
| `batch_import` | completed | 批量导入完成 | 成功导入 5 个文件，失败 1 个 |
| `batch_import` | failed | 批量导入失败 | 批量导入任务执行失败 |

批量导入的通知消息可从 `BatchImportResult` 生成，失败项过多时截断。

## 测试策略

### 后端

1. `TestSingleImportHandler_Execute_Success`：验证 handler 成功调用 `FileService.ImportFile`；
2. `TestSingleImportHandler_Resume_ReturnsError`：验证不可恢复；
3. `TestJobService_EmitsNotificationOnCompletion`：验证任务完成时发送 `notification:new` 事件；
4. `TestJobService_EmitsNotificationOnFailure`：验证任务失败时发送通知事件。

### 前端

1. `notifications store`：添加、标记已读、清空、未读数计算；
2. `NotificationPanel.vue`：空状态、列表渲染、点击跳转；
3. `ImportView.vue`：提交后显示成功提示，不调用 `router.push`。

## 演进预留（到方案 B）

后续如果要将通知持久化到后端，改动范围如下：

1. **后端**：新增 `Notification` 表、`NotificationRepository`、`NotificationService`、Wails 绑定（`ListNotifications`、`MarkNotificationRead`、`ClearNotifications`）；
2. **前端**：`notifications store` 的 `initialize()` 改为先调用后端拉取历史，再监听事件；
3. **事件格式不变**：`notification:new` 字段保持完全一致；
4. **ID 生成提前统一**：A+ 阶段由后端生成 ID，避免切换后冲突。

## 明确不做（当前阶段）

1. 不持久化通知，应用关闭后消息中心清空；
2. 不实现通知的音效、桌面级系统通知；
3. 不实现按类型过滤、搜索通知；
4. 不改动批量导入的核心逻辑，只复用其通知事件。
