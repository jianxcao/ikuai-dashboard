import * as echarts from 'echarts/core'
import { GraphChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { LabelLayout } from 'echarts/features'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  LineChart,
  PieChart,
  GraphChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  LabelLayout,
  CanvasRenderer,
])

export { echarts }
export type { EChartsCoreOption } from 'echarts/core'
