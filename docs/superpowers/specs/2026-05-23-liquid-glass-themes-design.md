# 液态玻璃双主题重构设计

## 背景

当前前端是一个基于 Vue 3 和 Vite 的爱快路由器监控面板。项目已经在 `frontend/src/assets/main.css` 中使用了暗色半透明玻璃风格，但整体仍是单主题实现，部分样式分散在页面和组件局部，缺少可复用、可扩展的液态玻璃设计系统。

已确认的重构方向是围绕两套相关主题改造 UI：

- `liquid-dark`：默认暗色液态玻璃主题，适合长时间监控场景。
- `liquid-light`：浅色液态玻璃主题，复用同一套组件语言，但调整光源、对比度和玻璃设计变量。

本次重构不改变后端 API、DTO、轮询逻辑和路由结构。

## 目标

- 提供至少两套可由用户切换的主题。
- 两套主题都要具备液态玻璃质感，而不是普通的 `blur()` 面板。
- 保持实时网络数据的可读性。
- 实现范围限定在前端。
- 沿用现有 Vue 3 组合式 API 写法。
- 不引入大型 UI 框架，不替换现有应用架构。

## 非目标

- 不修改后端。
- 不修改 API 响应结构。
- 不替换 ECharts、Axios、Vite 或 Vue。
- 不引入新的前端路由库。
- 不做超出主题一致性所需的信息架构大改。

## 主题架构

应用根节点通过 `data-theme` 标识当前主题：

```html
<div id="app" data-theme="liquid-dark">
```

主题切换器通过 Vue 状态更新该属性，并把用户选择持久化到 `localStorage`。

主题状态放在一个小型 composable 中：

- `frontend/src/composables/useTheme.ts`
- 暴露 `theme`、`setTheme`、`toggleTheme` 和 `themes`。
- 读取 `localStorage.getItem('ikuai_theme')`。
- 默认主题为 `liquid-dark`。
- 把当前主题应用到 `document.documentElement.dataset.theme`。

使用 `document.documentElement` 可以让主题变量对全局元素、ECharts 容器和未来可能出现的浮层都可用。

## 主题设计变量

全局 CSS 变量在 `frontend/src/assets/main.css` 中定义视觉系统。两套主题覆盖同一组设计变量名称。

核心设计变量分组：

- 背景：`--app-bg`、`--ambient-1`、`--ambient-2`、`--ambient-3`、`--mesh-opacity`。
- 文本：`--text-primary`、`--text-secondary`、`--text-tertiary`、`--text-muted`。
- 玻璃：`--glass-bg`、`--glass-bg-strong`、`--glass-bg-hover`、`--glass-border`、`--glass-highlight`、`--glass-shadow`、`--glass-inner-shadow`。
- 控件：`--control-bg`、`--control-bg-hover`、`--control-active`、`--control-border`。
- 状态：`--system-blue`、`--system-green`、`--system-orange`、`--system-red`，以及对应的 dim 变体。
- 布局：保留现有圆角变量，针对高密度控件补充更小圆角。

## 液态玻璃视觉语言

液态玻璃组件由多层效果叠加构成：

- 带饱和度的透明填充和背景模糊。
- 随主题变化的细边框。
- 模拟折射光的内部顶部高光。
- 用于抬升面板层级的柔和外阴影。
- 卡片和主要控件上的 `::before` 光泽层。
- 交互表面的轻微 hover 浮动。
- 能透过玻璃表面看到的背景 mesh 和细噪点。

设计必须保持控件行为可预期。动效要克制，不能影响实时指标扫描。

## 主题切换器

`AppHeader.vue` 右侧新增紧凑型分段主题控件。

要求：

- 能展示当前激活主题。
- 使用 `lucide-vue-next` 图标。
- 切换主题时不离开当前标签页。
- 刷新页面后保持用户选择。
- 桌面端和移动端都不能与导航标签页重叠。

页头布局需要更响应式：

- 桌面端：品牌、导航标签页、状态、主题切换器。
- 平板和移动端：必要时换行，而不是裁切文字或控件。

## 组件迁移范围

把现有 CSS 迁移到共享主题变量和液态玻璃工具类上：

- `frontend/src/assets/main.css`
  - 主题设计变量。
  - 全局背景。
  - 共享 `.glass-card`、`.glass-panel`、`.liquid-button`、`.theme-chip`、错误和加载工具类。
- `frontend/src/App.vue`
  - 保留全局背景元素，必要时调整 class 行为。
- `frontend/src/components/layout/AppLayout.vue`
  - 保留当前动态组件导航。
  - 加入主题状态，并传给 header。
- `frontend/src/components/layout/AppHeader.vue`
  - 新增主题切换器。
  - 优化响应式 header 布局。
- 监控卡片、表格、图表、LAN 客户端组件和搜索控件
  - 用变量替换硬编码颜色。
  - 在合适位置使用共享玻璃类。

`NetworkMap.vue`、`SecurityHub.vue`、`MultiWan.vue` 这类使用直接后端 URL 的视图不是本次 API 整理重点，但可见面板样式仍需接入主题系统。

## 数据流

监控数据流不变。

主题数据流：

```text
AppLayout -> useTheme -> document.documentElement[data-theme]
          -> AppHeader props/events -> 用户切换主题
          -> localStorage('ikuai_theme')
          -> CSS 变量更新所有表面
```

## 可访问性与响应式

- 主题切换按钮必须使用真实 `<button>`。
- 激活主题需要有 `aria-pressed="true"` 或等效状态。
- 两套主题的文字对比度都必须可用。
- 窄屏下页头控件不能重叠。
- 玻璃表面上的 `focus-visible` 状态必须清晰。
- 不使用基于 viewport 的字体缩放。

## 测试与验证

运行：

```sh
cd frontend
pnpm run build
```

再使用 mock 后端运行应用：

```sh
MOCK_MODE=true go run ./backend/cmd/server
cd frontend
pnpm dev
```

浏览器验证需要覆盖：

- 默认 `liquid-dark` 主题可正常渲染。
- 切换到 `liquid-light` 后界面无需刷新即可更新。
- 刷新后仍保持上一次选择的主题。
- 首页看板、局域网客户端和网络拓扑图都以玻璃表面渲染。
- 桌面和移动宽度下页头与控件没有明显重叠。

## 风险

- 组件 scoped CSS 中的硬编码颜色可能削弱主题一致性。
- ECharts 配置需要使用主题变量感知颜色，否则浅色主题下标签可能不可读。
- 部分异步视图使用 `axios.get('http://localhost:8080/...')` 直连后端，可能绕过 Vite proxy；除非视觉验证必须，不把本次重构扩大为 API 清理。

## 验收标准

- 应用有两套可选择主题：`liquid-dark` 和 `liquid-light`。
- 主题选择持久化到 `localStorage`。
- 主要表面具备液态玻璃样式：模糊、透明、折射高光和主题化背景光。
- 现有监控数据行为保持不变。
- `pnpm run build` 成功。
- 浏览器检查确认两套主题都能渲染，主要面板页面仍然可读。
