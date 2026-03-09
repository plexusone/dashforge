import { useMemo } from 'react'
import ReactECharts from 'echarts-for-react'
import type { Widget, ChartConfig } from '../../types/dashboard'
import type { EChartsOption } from 'echarts'

interface ChartWidgetProps {
  widget: Widget
}

// Sample data for preview when no data source is connected
const sampleData = {
  categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
  series1: [120, 200, 150, 80, 70, 110],
  series2: [60, 120, 90, 150, 200, 170]
}

export function ChartWidget({ widget }: ChartWidgetProps) {
  const config = widget.config as ChartConfig

  const options: EChartsOption = useMemo(() => {
    const baseOptions: EChartsOption = {
      animation: true,
      animationDuration: 300,
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: widget.title ? '15%' : '10%',
        containLabel: true
      },
      tooltip: {
        trigger: config.geometry === 'pie' ? 'item' : 'axis'
      }
    }

    // Add legend if enabled
    if (config.style?.showLegend !== false) {
      baseOptions.legend = {
        show: true,
        top: 0,
        orient: 'horizontal'
      }
    }

    // Add title if present
    if (widget.title) {
      baseOptions.title = {
        text: widget.title,
        left: 'center',
        textStyle: {
          fontSize: 14,
          fontWeight: 'normal'
        }
      }
    }

    // Build chart-specific options
    switch (config.geometry) {
      case 'bar':
        return {
          ...baseOptions,
          xAxis: {
            type: config.style?.horizontal ? 'value' : 'category',
            data: config.style?.horizontal ? undefined : sampleData.categories
          },
          yAxis: {
            type: config.style?.horizontal ? 'category' : 'value',
            data: config.style?.horizontal ? sampleData.categories : undefined
          },
          series: [{
            name: 'Series 1',
            type: 'bar',
            data: sampleData.series1,
            itemStyle: {
              borderRadius: [4, 4, 0, 0]
            }
          }]
        }

      case 'line':
        return {
          ...baseOptions,
          xAxis: {
            type: 'category',
            data: sampleData.categories,
            boundaryGap: false
          },
          yAxis: {
            type: 'value'
          },
          series: [{
            name: 'Series 1',
            type: 'line',
            data: sampleData.series1,
            smooth: config.style?.smooth || false
          }]
        }

      case 'area':
        return {
          ...baseOptions,
          xAxis: {
            type: 'category',
            data: sampleData.categories,
            boundaryGap: false
          },
          yAxis: {
            type: 'value'
          },
          series: [{
            name: 'Series 1',
            type: 'line',
            data: sampleData.series1,
            smooth: config.style?.smooth || false,
            areaStyle: {
              opacity: 0.3
            },
            stack: config.style?.stack ? 'total' : undefined
          }, {
            name: 'Series 2',
            type: 'line',
            data: sampleData.series2,
            smooth: config.style?.smooth || false,
            areaStyle: {
              opacity: 0.3
            },
            stack: config.style?.stack ? 'total' : undefined
          }]
        }

      case 'pie':
        return {
          ...baseOptions,
          series: [{
            name: 'Distribution',
            type: 'pie',
            radius: ['40%', '70%'],
            center: ['50%', '55%'],
            avoidLabelOverlap: false,
            itemStyle: {
              borderRadius: 4,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              show: config.style?.showLabels !== false,
              position: 'outside'
            },
            data: sampleData.categories.map((cat, i) => ({
              name: cat,
              value: sampleData.series1[i]
            }))
          }]
        }

      case 'scatter':
        return {
          ...baseOptions,
          xAxis: {
            type: 'value'
          },
          yAxis: {
            type: 'value'
          },
          series: [{
            name: 'Series 1',
            type: 'scatter',
            symbolSize: 10,
            data: sampleData.series1.map((v, i) => [v, sampleData.series2[i]])
          }]
        }

      case 'radar':
        return {
          ...baseOptions,
          radar: {
            indicator: sampleData.categories.map(cat => ({
              name: cat,
              max: 250
            }))
          },
          series: [{
            name: 'Series',
            type: 'radar',
            data: [
              {
                name: 'Series 1',
                value: sampleData.series1
              },
              {
                name: 'Series 2',
                value: sampleData.series2
              }
            ]
          }]
        }

      case 'funnel':
        return {
          ...baseOptions,
          series: [{
            name: 'Funnel',
            type: 'funnel',
            left: '10%',
            top: 60,
            bottom: 20,
            width: '80%',
            min: 0,
            max: 100,
            minSize: '0%',
            maxSize: '100%',
            sort: 'descending',
            gap: 2,
            label: {
              show: true,
              position: 'inside'
            },
            data: [
              { value: 100, name: 'Awareness' },
              { value: 80, name: 'Interest' },
              { value: 60, name: 'Consideration' },
              { value: 40, name: 'Intent' },
              { value: 20, name: 'Purchase' }
            ]
          }]
        }

      case 'gauge':
        return {
          ...baseOptions,
          series: [{
            name: 'Gauge',
            type: 'gauge',
            center: ['50%', '60%'],
            startAngle: 200,
            endAngle: -20,
            min: 0,
            max: 100,
            splitNumber: 10,
            progress: {
              show: true,
              width: 20
            },
            pointer: {
              show: true
            },
            axisLine: {
              lineStyle: {
                width: 20
              }
            },
            axisTick: {
              distance: -30,
              splitNumber: 5,
              lineStyle: {
                width: 2,
                color: '#999'
              }
            },
            splitLine: {
              distance: -35,
              length: 14,
              lineStyle: {
                width: 3,
                color: '#999'
              }
            },
            axisLabel: {
              distance: -20,
              color: '#999',
              fontSize: 12
            },
            detail: {
              valueAnimation: true,
              formatter: '{value}%',
              color: 'inherit',
              fontSize: 24
            },
            data: [{ value: 72, name: 'Progress' }]
          }]
        }

      case 'heatmap':
        return {
          ...baseOptions,
          xAxis: {
            type: 'category',
            data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
          },
          yAxis: {
            type: 'category',
            data: ['Morning', 'Afternoon', 'Evening', 'Night']
          },
          visualMap: {
            min: 0,
            max: 100,
            calculable: true,
            orient: 'horizontal',
            left: 'center',
            bottom: 10
          },
          series: [{
            name: 'Activity',
            type: 'heatmap',
            data: [
              [0, 0, 10], [0, 1, 30], [0, 2, 60], [0, 3, 20],
              [1, 0, 20], [1, 1, 50], [1, 2, 80], [1, 3, 15],
              [2, 0, 15], [2, 1, 45], [2, 2, 70], [2, 3, 25],
              [3, 0, 25], [3, 1, 55], [3, 2, 90], [3, 3, 30],
              [4, 0, 30], [4, 1, 65], [4, 2, 85], [4, 3, 20],
              [5, 0, 40], [5, 1, 35], [5, 2, 50], [5, 3, 40],
              [6, 0, 50], [6, 1, 25], [6, 2, 40], [6, 3, 45]
            ],
            label: {
              show: true
            },
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowColor: 'rgba(0, 0, 0, 0.5)'
              }
            }
          }]
        }

      case 'treemap':
        return {
          ...baseOptions,
          series: [{
            name: 'Categories',
            type: 'treemap',
            roam: false,
            data: [
              {
                name: 'Electronics',
                value: 100,
                children: [
                  { name: 'Phones', value: 40 },
                  { name: 'Laptops', value: 35 },
                  { name: 'Tablets', value: 25 }
                ]
              },
              {
                name: 'Clothing',
                value: 80,
                children: [
                  { name: 'Shirts', value: 30 },
                  { name: 'Pants', value: 25 },
                  { name: 'Shoes', value: 25 }
                ]
              },
              {
                name: 'Home',
                value: 60,
                children: [
                  { name: 'Furniture', value: 25 },
                  { name: 'Appliances', value: 20 },
                  { name: 'Decor', value: 15 }
                ]
              }
            ],
            label: {
              show: true,
              formatter: '{b}'
            },
            upperLabel: {
              show: true,
              height: 20
            },
            itemStyle: {
              borderColor: '#fff',
              borderWidth: 2
            },
            levels: [
              {
                itemStyle: {
                  borderWidth: 3,
                  borderColor: '#333',
                  gapWidth: 3
                }
              },
              {
                colorSaturation: [0.35, 0.5],
                itemStyle: {
                  borderWidth: 2,
                  gapWidth: 2,
                  borderColorSaturation: 0.6
                }
              }
            ]
          }]
        }

      case 'sankey':
        return {
          ...baseOptions,
          series: [{
            name: 'Flow',
            type: 'sankey',
            layout: 'none',
            emphasis: {
              focus: 'adjacency'
            },
            data: [
              { name: 'Source A' },
              { name: 'Source B' },
              { name: 'Source C' },
              { name: 'Process 1' },
              { name: 'Process 2' },
              { name: 'Output X' },
              { name: 'Output Y' }
            ],
            links: [
              { source: 'Source A', target: 'Process 1', value: 30 },
              { source: 'Source A', target: 'Process 2', value: 20 },
              { source: 'Source B', target: 'Process 1', value: 25 },
              { source: 'Source B', target: 'Process 2', value: 35 },
              { source: 'Source C', target: 'Process 2', value: 15 },
              { source: 'Process 1', target: 'Output X', value: 40 },
              { source: 'Process 1', target: 'Output Y', value: 15 },
              { source: 'Process 2', target: 'Output X', value: 30 },
              { source: 'Process 2', target: 'Output Y', value: 40 }
            ],
            lineStyle: {
              color: 'gradient',
              curveness: 0.5
            }
          }]
        }

      default:
        return {
          ...baseOptions,
          xAxis: {
            type: 'category',
            data: sampleData.categories
          },
          yAxis: {
            type: 'value'
          },
          series: [{
            type: 'bar',
            data: sampleData.series1
          }]
        }
    }
  }, [config, widget.title])

  // Default colors
  const theme = {
    color: config.style?.colors || [
      '#5470c6', '#91cc75', '#fac858', '#ee6666',
      '#73c0de', '#3ba272', '#fc8452', '#9a60b4'
    ]
  }

  return (
    <div className="w-full h-full p-2">
      <ReactECharts
        option={options}
        theme={theme}
        style={{ width: '100%', height: '100%' }}
        opts={{ renderer: 'svg' }}
        notMerge={true}
      />
    </div>
  )
}
