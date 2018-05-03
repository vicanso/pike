<template lang="pug">
.directorPage
  el-card.director(
    v-for='item in directors'
    :key='item.name'
  )
    h4
      .pullRight
        span(
          v-if='item.policy'
        ) policy : {{item.policy}}
        span priority : {{item.priority}}
      | {{item.name}}
    h5 backends
    ul
      li(
        v-for='backend in item.backends'
        :key='backend.backend'
      )
        | {{backend.backend}}
        span.mleft5
          i.el-icon-circle-check-outline.healthy(
            v-if='backend.status === "healthy"'
          )
          i.el-icon-circle-close-outline.sick(
            v-else
          )
    
    div(
      v-if='item.hosts && item.hosts.length'
    )
      h5 hosts
      ul
        li(
          v-for='host in item.hosts'
          :key='host'
        ) {{host}}
        
    div(
      v-if='item.prefixs && item.prefixs.length'
    )
      h5 prefixs
      ul
        li(
          v-for='prefix in item.prefixs'
          :key='prefix'
        ) {{prefix}}

</template>

<script>
import {mapActions, mapState} from 'vuex';
export default {
  data() {
    return {};
  },
  methods: {
    ...mapActions(['getDirectors']),
  },
  computed: {
    ...mapState({
      directors: ({pike}) => pike.directors,
    }),
  },
  async beforeMount() {
    const close = this.$loading();
    try {
      await this.getDirectors()
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

.directorPage
  width: 1000px
  margin: auto
  padding-top: 20px
  .director
    margin-bottom: 20px
    h4
      margin: 0
      padding: 0
      line-height: 3em
      font-size: 18px
      span
        margin-left: 10px
        font-size: 12px
        color: $COLOR_DARK_GRAY
        font-weight: 600
    h5
      margin: 0
      padding: 10px
      background-color: $COLOR_DARY_WHITE
    ul
      margin: 20px 0
      padding: 0
      padding-left: 20px
      list-style: inside
      li
        line-height: 1.5em
    p
      margin: 20px
    .healthy
      color: $COLOR_BLUE
    .sick
      color: $COLOR_RED
</style>

