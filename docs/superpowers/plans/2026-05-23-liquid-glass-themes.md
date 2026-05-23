# 液态玻璃双主题实现计划

> **给自动化执行代理：** 必须使用子技能：优先使用 `superpowers:subagent-driven-development`，也可以使用 `superpowers:executing-plans` 按任务实现本计划。步骤使用复选框（`- [ ]`）语法跟踪。

**目标：** 重构 Vue 前端，支持两套可选择的液态玻璃主题：`liquid-dark` 和 `liquid-light`，并持久化用户选择，同时保持监控面板可读。

**架构：** 新增一个聚焦的主题组合函数，负责持久化主题状态并应用 `document.documentElement.dataset.theme`。把视觉基础设施迁移到 CSS 设计变量和共享液态玻璃工具类，再逐步迁移布局、监控、局域网和关键异步视图。后端 API、数据轮询和应用导航保持不变。

**技术栈：** Vue 3 组合式 API、Vite、TypeScript、Vue 作用域 CSS、CSS 变量、`lucide-vue-next`、ECharts。

---

## 文件结构

- 新建 `frontend/src/composables/useTheme.ts`：主题名称、持久化和 DOM 数据集应用。
- 修改 `frontend/src/assets/main.css`：两套主题设计变量和全局液态玻璃工具类。
- 修改 `frontend/src/App.vue`：保留全局氛围背景，并暴露液态背景层。
- 修改 `frontend/src/components/layout/AppLayout.vue`：初始化主题状态并传给页头组件。
- 修改 `frontend/src/components/layout/AppHeader.vue`：新增可访问的主题切换器和响应式液态页头。
- 修改 `frontend/src/components/monitor` 下的监控组件：用主题设计变量替换硬编码颜色，强化液态表面。
- 修改 `frontend/src/components/lan` 下的局域网组件：让搜索、列表和客户端卡片接入主题设计变量与响应式行为。
- 修改 `frontend/src/views/NetworkMap.vue`、`SecurityHub.vue`、`MultiWan.vue`：让可见面板主题化，并保持图表标签可读。

## 任务 1：新增主题状态

**文件：**
- 新建：`frontend/src/composables/useTheme.ts`

- [ ] **步骤 1：创建 composable**

```ts
import { computed, readonly, ref, watchEffect } from 'vue'

export type AppTheme = 'liquid-dark' | 'liquid-light'

export const themes: Array<{ id: AppTheme; label: string; shortLabel: string }> = [
  { id: 'liquid-dark', label: '液态暗色', shortLabel: '暗色' },
  { id: 'liquid-light', label: '液态浅色', shortLabel: '浅色' },
]

const STORAGE_KEY = 'ikuai_theme'
const fallbackTheme: AppTheme = 'liquid-dark'
const theme = ref<AppTheme>(readInitialTheme())

function isTheme(value: string | null): value is AppTheme {
  return value === 'liquid-dark' || value === 'liquid-light'
}

function readInitialTheme(): AppTheme {
  if (typeof window === 'undefined') return fallbackTheme
  const saved = window.localStorage.getItem(STORAGE_KEY)
  return isTheme(saved) ? saved : fallbackTheme
}

function applyTheme(value: AppTheme) {
  if (typeof document === 'undefined') return
  document.documentElement.dataset.theme = value
}

function setTheme(value: AppTheme) {
  theme.value = value
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, value)
  }
}

function toggleTheme() {
  setTheme(theme.value === 'liquid-dark' ? 'liquid-light' : 'liquid-dark')
}

watchEffect(() => applyTheme(theme.value))

export function useTheme() {
  return {
    theme: readonly(theme),
    themeLabel: computed(() => themes.find((item) => item.id === theme.value)?.label ?? '液态暗色'),
    themes,
    setTheme,
    toggleTheme,
  }
}
```

- [ ] **步骤 2：运行类型检查**

运行：`cd frontend && pnpm run type-check`

预期：`vue-tsc --build` 以退出码 0 结束。

## 任务 2：用液态主题设计变量替换全局样式

**文件：**
- 修改：`frontend/src/assets/main.css`
- 修改：`frontend/src/App.vue`

- [ ] **步骤 1：替换全局 CSS 为变量化液态玻璃基础样式**

实现：

- `:root` 默认暗色设计变量。
- `:root[data-theme='liquid-dark']` 和 `:root[data-theme='liquid-light']` 覆盖。
- `.glass-card`、`.glass-panel`、`.liquid-button`、`.theme-chip`、`.error-banner`、`.loading-state`、`.spinner`。
- `.app-bg-mesh::before` 和 `.app-bg-mesh::after` 用于环境光和颗粒层。

CSS 必须包含设计规格中的所有设计变量名称，包括 `--glass-highlight`、`--glass-inner-shadow`、`--control-bg` 和 `--text-muted`。

- [ ] **步骤 2：更新 `App.vue` 背景结构**

保持结构：

```vue
<template>
  <div class="app-bg-mesh" aria-hidden="true"></div>
  <AppLayout />
</template>
```

不修改数据行为。

- [ ] **步骤 3：运行构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。Vite chunk-size warning 可接受。

## 任务 3：新增页头主题切换器

**文件：**
- 修改：`frontend/src/components/layout/AppLayout.vue`
- 修改：`frontend/src/components/layout/AppHeader.vue`

- [ ] **步骤 1：在 layout 中接入主题状态**

新增：

```ts
import { useTheme } from '@/composables/useTheme'

const { theme, themes, setTheme } = useTheme()
```

把 `theme`、`themes` 和 `setTheme` 传给 `AppHeader`。

- [ ] **步骤 2：新增带类型的页头 props 和主题事件**

`AppHeader.vue` 接收：

```ts
import type { AppTheme } from '@/composables/useTheme'

const props = defineProps<{
  currentTab?: string
  connected?: boolean
  theme: AppTheme
  themes: Array<{ id: AppTheme; label: string; shortLabel: string }>
}>()

const emit = defineEmits<{
  (event: 'tab-change', value: string): void
  (event: 'theme-change', value: AppTheme): void
}>()
```

- [ ] **步骤 3：实现主题切换按钮**

每个主题按钮必须是 `<button type="button">`，并包含：

```vue
:aria-pressed="theme === item.id"
@click="emit('theme-change', item.id)"
```

使用 `lucide-vue-next` 中的 `Moon` 和 `SunMedium` 图标。

- [ ] **步骤 4：让页头响应式**

CSS 需要支持窄屏换行：

```css
.nav-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
```

移动宽度下标签页可以横向滚动或换行，但不能与状态和主题控件重叠。

- [ ] **步骤 5：运行构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。

## 任务 4：迁移监控看板表面

**文件：**
- 修改：`frontend/src/views/MonitorInterface.vue`
- 修改：`frontend/src/components/monitor/SummaryCards.vue`
- 修改：`frontend/src/components/monitor/TrafficTables.vue`
- 修改：`frontend/src/components/monitor/TrafficChart.vue`

- [ ] **步骤 1：在视图中使用共享错误和加载类**

移除已被全局工具类覆盖的重复页面局部错误和加载 CSS。只保留视图专属间距。

- [ ] **步骤 2：更新摘要卡片样式**

卡片应使用：

```css
background: var(--glass-bg);
border: 1px solid var(--glass-border);
box-shadow: var(--glass-shadow), var(--glass-inner-shadow);
```

指标图标背景使用状态弱化变量，例如 `--system-green-dim`。

- [ ] **步骤 3：更新表格样式**

尽量用 class 替换内联硬编码颜色：

```css
.metric-up { color: var(--system-green); }
.metric-down { color: var(--system-blue); }
.table-muted { color: var(--text-secondary); }
```

表头和行 hover 背景必须使用 `--table-header-bg` 和 `--table-row-hover`。

- [ ] **步骤 4：让图表感知主题**

在 `TrafficChart.vue` 中通过 `getComputedStyle(document.documentElement)` 读取 CSS 变量。

以下内容使用变量值：

- 提示框背景和文字。
- 坐标轴标签。
- 分割线。
- 上传/下载线条颜色。

当主题变化时，通过观察 `document.documentElement.dataset.theme` 或重新生成 option 更新图表。

- [ ] **步骤 5：运行构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。

## 任务 5：迁移 LAN 设备表面

**文件：**
- 修改：`frontend/src/views/MonitorLan.vue`
- 修改：`frontend/src/components/lan/SearchBar.vue`
- 修改：`frontend/src/components/lan/ClientList.vue`
- 修改：`frontend/src/components/lan/ClientCard.vue`

- [ ] **步骤 1：保持 LAN 行为不变**

不修改 `useLanClients`、排序 key、轮询间隔或搜索 debounce。

- [ ] **步骤 2：更新搜索栏**

使用全局 `glass-card` 液态样式和主题变量焦点环：

```css
box-shadow: 0 0 0 3px var(--system-blue-dim), var(--glass-shadow);
```

- [ ] **步骤 3：更新列表和客户端卡片布局**

保留桌面列宽，并增加移动端兜底布局：

```css
@media (max-width: 920px) {
  .list-header { display: none; }
  .client-card { align-items: flex-start; flex-direction: column; }
}
```

- [ ] **步骤 4：替换硬编码白色透明表面**

使用 `--control-bg`、`--control-bg-hover`、`--control-border`、`--text-secondary` 和状态弱化变量。

- [ ] **步骤 5：运行构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。

## 任务 6：对齐次级视图和图表

**文件：**
- 修改：`frontend/src/views/NetworkMap.vue`
- 修改：`frontend/src/views/SecurityHub.vue`
- 修改：`frontend/src/views/MultiWan.vue`

- [ ] **步骤 1：更新面板样式到共享变量**

使用 `.glass-panel`、`--glass-border`、`--control-bg` 和 `--text-*` 变量。

- [ ] **步骤 2：让网络拓扑图中的 ECharts 在两套主题下可读**

使用 `echarts.init(chartRef.value)`，不再固定暗色主题，并从 CSS 变量设置标签和文字颜色。

- [ ] **步骤 3：除非构建需要，不改 API 调用**

不把本任务扩展成 API 清理。现有 `axios.get('http://localhost:8080/...')` 调用可以保留。

- [ ] **步骤 4：运行构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。

## 任务 7：运行时验证

**文件：**
- 不修改源码，除非验证暴露了明确缺陷。

- [ ] **步骤 1：用 mock 模式启动后端**

运行：`MOCK_MODE=true go run ./backend/cmd/server`

预期：后端监听 `http://localhost:8080`。

- [ ] **步骤 2：启动前端**

运行：`cd frontend && pnpm dev -- --host 127.0.0.1`

预期：Vite 服务运行在 `http://127.0.0.1:5173`。

- [ ] **步骤 3：浏览器验证**

检查：

- `liquid-dark` 下的首页看板。
- 切换到 `liquid-light`。
- 刷新后确认主题持久化。
- 两套主题下的局域网客户端页面。
- 至少一个主题下的网络拓扑图。
- 桌面和移动宽度下页头与控件没有重叠。

- [ ] **步骤 4：最终构建**

运行：`cd frontend && pnpm run build`

预期：构建以退出码 0 结束。
