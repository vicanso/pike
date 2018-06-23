<template lang="pug">
.fetchingPage
  el-input(
    placeholder="please input keyword (support regexp)"
    clearable
    v-model='keyword'
  )
    template(
      slot="prepend"
    ) keyword
    template(
      slot="append"
    ) {{(fetchingList && fetchingList.length) || 0}} Fetching
  el-table.mtop10(
    :data='fetchingList'
    stripe
  )
    el-table-column(
      prop='key'
      label='Key'
      sortable
    )
    el-table-column(
      prop='timeConsuming'
      label='Time Consuming'
      width='180'
      sortable
    )
      template(
        slot-scope='scope'
      )
        span {{scope.row.timeConsumingDesc}}
    el-table-column(
      prop='createdAt'
      label='CreatedAt'
      width='160'
      sortable
    )
</template>
<script src="./fetching.js"></script>
<style lang="sass" scoped>
@import '../../variables'
.fetchingPage
  padding: 20px
</style>
