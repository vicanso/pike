<template lang="pug">
.performancePage
  .bkz
    h3.bla.blb QUICK PERFORMANCE

  el-row(
    v-if='performances'
  )
    el-col(
      :span='8'
      v-for='item in performanceItems'
      :key='item.key'
    )
      .performance(
        :class='item.cls'
      )
        h4 {{item.name}}
        h5 {{getCount(item.key)}}
        Chart.chartView(
          :name='item.key'
          :data='performances'
        )
        p.font12 {{item.desc}}

</template>

<script src="./performance.js"></script>

<style lang="sass" scoped>
@import '../../variables'

.performancePage
  padding: 15px
.performance
  margin: 15px
  border-radius: 3px
  padding: 25px
  h4
    margin: 0
    padding: 0 
    font-weight: normal
    color: rgba(255, 255, 255, 0.65)
    text-align: left
    text-transform: uppercase
    font-size: 14px 
    letter-spacing: .15em
  h5
    font-size: 2rem
    line-height: 1.5em
    margin: 0
    border-bottom: 1px solid rgba(255, 255, 255, 0.2)
    color: rgba(255, 255, 255, 0.65)
  .chartView
    width: 100%
    height: 180px
  p
    margin: 0
    padding: 5px 0
    color: $COLOR_WHITE
    white-space: nowrap
    text-overflow: ellipsis
    overflow: hidden
</style>
