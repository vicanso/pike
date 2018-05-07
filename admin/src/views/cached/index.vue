<template lang="pug">
.cachedPage
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
    ) {{(cachedList && cachedList.length) || 0}} Cache
  el-table.mtop10(
    :data='cachedList'
    stripe
  )
    el-table-column(
      prop='key'
      label='Key'
      sortable
    )
    el-table-column(
      prop='ttl'
      label='TTL'
      width='100'
      sortable
    )
    el-table-column(
      prop='expiredSeconds'
      label='Expired'
      width='100'
      sortable
    )
      template(
        slot-scope='scope'
      )
        span {{scope.row.expiredDesc}}
    el-table-column(
      prop='createdAt'
      label='CreatedAt'
      width='160'
      sortable
    )
    el-table-column(
      label='OP'
      width='80'
    )
      template(
        slot-scope='scope'
      )
        a.clear(
          href='javascript:;'
          @click='clear(scope.row)'
        ) clear

</template>

<script src="./cache.js"></script>
<style lang="sass" scoped>
@import '../../variables'
.cachedPage
  padding: 20px
  .clear
    text-decoration: none
    color: $COLOR_BLUE
</style>

