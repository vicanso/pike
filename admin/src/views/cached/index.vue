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

<script>
import {mapActions, mapState} from 'vuex';
import _ from 'lodash';

export default {
  data() {
    return {
      keyword: '',
    };
  },
  methods: {
    ...mapActions(['getCached', 'clearCached']),
    async clear(item) {
      const close = this.$loading();
      try {
        await this.clearCached(item.key);
      } catch (err) {
        this.$error(err);
      } finally {
        close();
      }
    },
  },
  computed: {
    ...mapState({
      cacheds: ({pike}) => pike.cacheds,
    }),
    cachedList: function() {
      const {
        cacheds,
        keyword,
      } = this;
      const reg = new RegExp(keyword, 'gi');
      return _.filter(cacheds, (item) => {
        if (!keyword) {
          return true;
        }
        return reg.test(item.key);
      })
    },
  },
  async beforeMount() {
    const close = this.$loading();
    try {
      await this.getCached()
    } catch (err) {
      this.$error(err)
    } finally {
      close();
    }
  },
}
</script>
<style lang="sass" scoped>
@import '../../variables'
.cachedPage
  padding: 20px
  .clear
    text-decoration: none
    color: $COLOR_BLUE
</style>

