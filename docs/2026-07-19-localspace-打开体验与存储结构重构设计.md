# LocalSpace 打开体验与存储结构重构设计

日期：2026-07-19

## 背景

当前 LocalSpace 在“文件打开体验”上存在两个明显阻力：

1. 软件内不支持为不同类型文件设置默认打开软件
2. 导入后的文件虽然已经按主目录和 `fileType` 落盘，但同一类型文件仍然堆在同一层目录下

这会导致一个典型问题：

- 用户在 LocalSpace 中找到目标视频
- 点击“打开所在目录”
- 进入资源管理器
- 右键文件
- 再选择播放器

如果文件规模继续增加，尤其是电视剧、课程视频、素材合集等内容增多后，这条链路会越来越重。

因此，这次设计的目标不是单独补一个“默认软件设置”，而是把“打开方式”和“存储结构”作为一个连续体验来重构。

## 问题判断

### 当前已经具备的基础

从现有实现看：

- 导入时已经会把文件移动到软件管理目录
- 已经存在“主目录 -> 子目录(fileType)”的基本存储模型
- `OpenFile` 当前仅调用系统默认程序打开
- `OpenFileLocation` 已支持直接在系统文件管理器中定位文件

对应关键实现主要在：

- `app/app.go`
- `app/services/file_service.go`
- `app/services/storage_service.go`
- `app/repositories/config_repository.go`

### 当前体验上的真实瓶颈

当前问题不是“不能打开文件”，而是“打开目标文件的路径太绕”。

根因分成两层：

#### 第一层：打开方式不可控

- 视频想用播放器 A
- 文档想用阅读器 B
- 图片想用看图工具 C
- 但系统默认软件未必符合用户在资料管理场景中的习惯

所以用户需要绕到目录里右键“打开方式”。

#### 第二层：物理目录不够可读

虽然已有 `master/fileType/` 这一层，但视频类、文档类一多，同类文件仍然会堆在一个目录下。

对于“电视剧 / 课程 / 项目资料 / 某个合集”这种场景，仅按文件类型分层仍然不够。

## 产品目标

本次设计目标：

1. 用户在 LocalSpace 内可以一键用自己指定的软件打开文件
2. 用户不必再为了“切换播放器”而退回资源管理器
3. 软件管理目录在外部文件管理器中也应保持可读
4. 存储结构要支持未来大量文件，而不是只适合当前规模
5. 方案应尽量复用现有主目录与子目录模型，避免推翻式重做

## 非目标

本次不包含：

1. 不做“自动识别电视剧季/集信息”的复杂媒体库能力
2. 不做在线播放器或内置播放器
3. 不做跨平台复杂的应用注册中心
4. 不在本次重构里直接实现自动目录监听导入

## 总体方案

本次采用“两阶段、一体化”的方案：

### 阶段一：先解决怎么打开

新增“打开方式设置”，允许用户为文件类型或扩展名指定默认打开软件。

核心收益：

- 立即缩短打开链路
- 即使目录结构暂时未升级，使用体验也会明显改善

### 阶段二：再解决文件怎么摆

在现有 `主目录 -> 文件类型子目录` 的基础上，升级为“可配置的二级结构”。

核心收益：

- 外部目录更可读
- 方便用户在 LocalSpace 之外协作和使用文件
- 为未来批量导入、系列管理、规则模板打基础

## 信息架构设计

建议在设置页新增两个设置区域：

### 1. 打开方式

用于控制 LocalSpace 内部“打开文件”的行为。

### 2. 存储规则

用于控制新导入文件的物理落盘结构。

这样能让用户理解：

- “打开方式”决定双击后用什么软件打开
- “存储规则”决定文件被整理到什么目录结构里

## 功能设计一：默认打开软件

### 用户需求

用户希望：

- 视频默认用固定播放器打开
- 文档默认用固定阅读器打开
- 图片默认用固定看图软件打开
- 未配置时仍然可以走系统默认

### 配置层级

建议采用两级优先级：

1. 扩展名级别
2. 文件类型级别
3. 系统默认兜底

优先级规则：

1. 若当前扩展名存在专属配置，则优先使用
2. 否则查找文件类型配置
3. 若仍未配置，则回退到系统默认打开

示例：

- `.mkv` -> `PotPlayer`
- `video` -> `VLC`
- `document` -> `Typora`

这样 `.mkv` 会优先走 `PotPlayer`，而 `.mp4` 仍可走 `video` 类型级别配置。

### 设置页交互

建议新增一个组件，例如：

- `OpenWithConfig.vue`

界面包含：

1. 文件类型默认打开软件列表
2. 扩展名覆盖规则列表
3. “恢复系统默认”按钮
4. “测试打开方式”辅助按钮

#### 文件类型设置项

建议至少支持：

- 视频
- 音频
- 图片
- 文档
- 压缩包
- 安装包
- 其他

每项包含：

- 当前配置的软件路径
- 浏览选择可执行文件
- 清空配置

#### 扩展名覆盖规则

支持用户手动新增：

- 扩展名，如 `.mkv`
- 软件路径，如 `C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe`

每条规则支持：

- 编辑
- 删除
- 排序不是必须，按精确扩展名匹配即可

### 文件页交互

建议对文件卡片和列表交互做如下调整：

#### 默认行为

- 双击文件：使用 LocalSpace 配置的软件打开

#### 菜单项建议

在现有三点菜单中补充：

- 用默认软件打开
- 用系统默认打开
- 选择其他软件打开
- 打开所在目录

排序建议：

1. 文件详情
2. 用默认软件打开
3. 用系统默认打开
4. 选择其他软件打开
5. 打开所在目录
6. 重命名
7. 编辑标签和描述
8. 删除文件

其中：

- “用默认软件打开” 走 LocalSpace 配置规则
- “用系统默认打开” 直接走当前已有系统打开逻辑
- “选择其他软件打开” 可作为后续增强项，允许临时选择一次可执行文件

## 功能设计二：存储结构升级

### 当前结构

当前导入逻辑实际更接近：

`主目录 / fileType / 文件名`

例如：

`H:\LocalSpace\video\episode01.mkv`

### 问题

这比“所有文件全堆一个目录”已经好很多，但还不够：

- 同类文件多了后仍会非常拥挤
- 外部浏览时不利于按合集理解内容
- 电视剧、课程、项目资料等天然具有分组属性

### 目标结构

建议升级为：

`主目录 / fileType / 分组目录 / 文件名`

例如：

- `H:\LocalSpace\video\甄嬛传\第01集.mkv`
- `H:\LocalSpace\document\项目A\需求说明.docx`
- `H:\LocalSpace\image\旅行-日本-2026\IMG001.jpg`

### 分组目录来源

建议先采用“用户可控优先”的方案，而不是全自动猜测。

分组目录值来源优先级：

1. 导入时用户手动填写的“系列名/项目名/合集名”
2. 若为空，则尝试从标签或关键词中选择一个稳定字段
3. 若仍为空，则使用默认分组，例如 `_unsorted`

推荐第一版只做：

1. 显式新增导入字段 `collectionName`
2. 为空时落到 `_unsorted`

这是最稳妥的方案，避免一开始就让 AI 或规则猜错目录结构。

## 存储规则设计

### 新增配置项

建议引入新的存储规则配置对象，例如：

```json
{
  "strategy": "type_collection",
  "unsortedFolderName": "_unsorted",
  "sanitizeFolderName": true
}
```

字段说明：

- `strategy`
  - `type_only`
  - `type_collection`
- `unsortedFolderName`
  - 未指定合集名时使用的目录名
- `sanitizeFolderName`
  - 是否自动清理非法路径字符

第一版建议只支持两种策略：

1. `type_only`
   - `主目录 / fileType / 文件名`
2. `type_collection`
   - `主目录 / fileType / collectionName / 文件名`

这样既兼容当前逻辑，也方便逐步切换。

### 为什么不建议一开始支持太多模板

比如这类模板先不要做：

- `主目录 / 年份 / 文件类型 / 项目`
- `主目录 / 项目 / 文件类型 / 月份`

因为虽然灵活，但会明显增加：

- 设置复杂度
- 路径冲突风险
- 迁移复杂度
- 后续维护成本

第一版把“按类型”和“按类型 + 合集”做稳，已经足够解决你当前提到的核心问题。

## 数据模型设计

### 1. 文件模型新增字段

建议在 `files` 表与 `models.File` 中新增：

- `collection_name`

用途：

- 记录文件所属合集 / 系列 / 项目
- 参与落盘路径生成
- 为未来批量导入、系列浏览、筛选视图做准备

建议在 Go 模型中新增：

```go
CollectionName string `json:"collectionName"`
```

### 2. 打开方式配置

不建议单独新建数据库表，第一版直接复用 `configs` 键值存储即可。

建议新增配置 key：

- `open_with_config`
- `storage_layout_config`

配置值使用 JSON 字符串保存。

原因：

- 当前项目已经有稳定的 `configs` 表
- 这两个配置都更偏“全局设置对象”
- 不必额外为少量结构化配置开表

### 3. 配置结构建议

#### `open_with_config`

```json
{
  "byFileType": {
    "video": "C:\\Program Files\\DAUM\\PotPlayer\\PotPlayerMini64.exe",
    "document": "C:\\Program Files\\Typora\\Typora.exe"
  },
  "byExtension": {
    ".mkv": "C:\\Program Files\\DAUM\\PotPlayer\\PotPlayerMini64.exe"
  }
}
```

#### `storage_layout_config`

```json
{
  "strategy": "type_collection",
  "unsortedFolderName": "_unsorted",
  "sanitizeFolderName": true
}
```

## 后端接口设计

### 打开方式相关

建议新增专用配置接口，而不是继续全部走通用 key-value。

新增：

- `GetOpenWithConfig()`
- `UpdateOpenWithConfig(config)`
- `OpenFileWithPreferredApp(id uint)`
- `OpenFileWithSystemDefault(id uint)`
- `SelectExecutable()`

说明：

- `OpenFileWithPreferredApp` 用于软件内主打开动作
- `OpenFileWithSystemDefault` 保留当前能力
- `SelectExecutable` 用于在设置页选择 `.exe`

其中：

- 当前 `OpenFile` 建议后续语义调整为“优先打开”
- 或者新增方法后逐步把前端调用切过去

### 存储规则相关

新增：

- `GetStorageLayoutConfig()`
- `UpdateStorageLayoutConfig(config)`

### 导入相关

导入接口需要补充：

- `collectionName`

例如将：

- `ImportFileWithKeywords(filePath, fileName, description, tags, keywords)`

升级为：

- `ImportFileWithMetadata(filePath, fileName, description, tags, keywords, collectionName)`

如果暂时不想大改接口，也可以新增一个版本化接口，保留旧接口兼容。

## 后端行为设计

### 1. 打开文件逻辑

新增一个统一解析函数，例如：

- `ResolveOpenCommand(filePath, fileType, fileSubType, config)`

逻辑：

1. 先读取扩展名
2. 检查是否命中扩展名级配置
3. 否则检查文件类型级配置
4. 若命中，则以指定程序 + 文件路径执行
5. 若未命中，则回退系统默认

Windows 下建议执行形式：

- `exec.Command(appPath, filePath)`

需要注意：

- 路径中包含空格
- 中文路径
- 不同播放器是否要求额外参数

第一版先只支持“程序路径 + 文件路径”即可，不做复杂启动参数模板。

### 2. 目录路径生成逻辑

建议新增统一函数，例如：

- `GetStoragePathForImport(masterID, fileType, collectionName, fileName)`

逻辑：

1. 读取 `storage_layout_config`
2. 根据策略决定目标目录
3. 若需要合集层级，则计算 `collectionSegment`
4. 清洗非法字符
5. 返回完整路径

示例：

- `type_only` -> `master/video/file.mkv`
- `type_collection` -> `master/video/甄嬛传/file.mkv`
- 无合集名 -> `master/video/_unsorted/file.mkv`

### 3. 文件名与目录名清洗

建议统一做 sanitize：

- 去除前后空格
- 替换 Windows 非法字符
- 避免结尾为点和空格

例如：

- `: * ? " < > | / \`

应替换为 `_`

## 前端设计

### 设置页

建议在 [frontend/src/views/SettingsView.vue](H:\mySpace\myGoSpace\localSpace\frontend\src\views\SettingsView.vue:1) 中新增两个卡片区域：

1. 打开方式
2. 存储规则

#### 打开方式卡片

建议内容：

- 文件类型默认软件表单
- 扩展名覆盖规则表单
- 软件路径浏览按钮
- 清空回退系统默认按钮

#### 存储规则卡片

建议内容：

- 存储结构策略选择
  - 按类型
  - 按类型 + 合集
- 未分组目录名设置
- 合法化目录名开关
- 说明文案

### 导入页

建议在 [frontend/src/components/FileMetaForm.vue](H:\mySpace\myGoSpace\localSpace\frontend\src\components\FileMetaForm.vue:1) 中新增：

- `collectionName`

字段文案建议：

- 系列 / 项目 / 合集

提示文案：

- 用于决定文件落到哪个分组目录，例如剧名、课程名、项目名

第一版不必强制必填。

### 文件卡片

建议在 [frontend/src/components/FileCard.vue](H:\mySpace\myGoSpace\localSpace\frontend\src\components\FileCard.vue:1) 调整菜单动作：

- 把当前默认点击行为逐步收敛到“优先打开”
- 菜单中保留“系统默认打开”
- 打开所在目录继续保留

## 迁移策略设计

这是本次最关键的风险控制点之一。

### 核心原则

存储规则升级不应强迫用户立即迁移旧文件。

建议采用：

- 新规则只影响新导入文件
- 旧文件保持原路径可正常使用
- 后续再提供手动迁移工具

### 原因

如果一上来就自动迁移老文件，会带来：

- 路径变更风险
- 外部引用失效风险
- 大量文件移动耗时
- 中断时状态不一致

### 第一版方案

第一版只做：

1. 新配置生效后，新导入文件按新规则落盘
2. 旧文件保持原位置
3. 打开逻辑完全兼容旧路径
4. `RefreshFile` 继续以真实路径为准

### 第二版迁移工具

后续可单独做：

- “整理现有文件”工具

能力包括：

- 预览将移动哪些文件
- 预览目标路径
- 跳过冲突
- 批量更新数据库路径
- 回滚失败项

但不建议在本次一起做。

## 实施顺序建议

### Phase 1：默认打开软件

1. 新增 `OpenWithConfig` 配置结构
2. 新增配置读写接口
3. 新增 `OpenFileWithPreferredApp`
4. 设置页增加“打开方式”模块
5. 文件卡片菜单接入“默认打开 / 系统默认打开”

这是最先能缓解你当前痛点的部分。

### Phase 2：存储规则配置

1. 新增 `StorageLayoutConfig`
2. 新增配置读写接口
3. 文件模型新增 `collectionName`
4. 导入表单新增“系列 / 项目 / 合集”
5. 落盘路径生成改为走新规则

### Phase 3：补充浏览体验

1. 文件详情页展示 `collectionName`
2. 支持按合集筛选
3. 支持从文件页直接查看同合集内容

## 风险与控制

### 风险 1：指定软件路径失效

情况：

- 用户移动了播放器安装位置
- 软件被卸载

控制方式：

- 打开失败时提示“指定软件不可用，是否改用系统默认打开”
- 设置页可标记配置状态异常

### 风险 2：合集名导致非法路径

情况：

- 用户输入了 Windows 不允许的字符

控制方式：

- 导入时前端提示
- 后端再次 sanitize
- 真正落盘前统一处理

### 风险 3：改了存储规则后用户误以为旧文件也会自动整理

控制方式：

- 设置页明确提示“仅影响新导入文件”
- 后续再单独提供“迁移现有文件”工具

### 风险 4：打开行为语义切换导致混淆

情况：

- 现在用户理解“打开文件”就是系统默认
- 以后会变成 LocalSpace 优先打开

控制方式：

- 菜单里同时保留：
  - 用默认软件打开
  - 用系统默认打开
- 设置页有清晰说明

## 最终结论

这次重构应该明确拆成两个层次：

1. 先解决“怎么打开”
2. 再解决“文件怎么摆”

从用户价值看：

- 默认打开软件，解决的是“打开慢、步骤绕”
- 存储结构升级，解决的是“目录乱、规模化后难管理”

结合你当前的真实痛点，推荐开发顺序是：

1. 默认打开软件配置
2. 文件页接入优先打开逻辑
3. 存储规则配置
4. 导入时增加合集字段
5. 新文件按“类型 + 合集”结构落盘

这样既能快速缓解当前使用痛点，也能让 LocalSpace 为更大规模的资料管理做好结构准备。
