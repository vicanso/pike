<template lang="pug">
.mainHeader
  //- .logo.clearfix.pullLeft
  //-   img.pullLeft(
  //-     :src='logo'
  //-   )
  //-   | Pike 
  //-   span(
  //-     v-if='statsInfo'
  //-   ) ({{statsInfo.version}})
  .pullRight.startedAt(
    v-if='statsInfo'
  ) {{statsInfo.startedAt}}
  ul.functions.pullLeft
    li(
      :class=`{
        active: currentRoute == item.route
      }`
      v-for='item in functions'
    ) {{item.name}}
</template>
<style lang="sass" scoped>
@import "../variables";
.mainHeader
  position: fixed
  left: 0
  top: 0
  right: 0
  padding-left: 30px
  height: $MAIN_HEADER_HEIGHT
  background-color: $COLOR_BLACK
  border-bottom: $GRAY_BORDER
  z-index: 9
  line-height: $MAIN_HEADER_HEIGHT
  .logo
    width: 150px 
    color: $COLOR_WHITE
    font-size: 18px
    span
      font-size: 14px
    img
      $imgHeight: 30px
      display: block
      width: $imgHeight
      height: $imgHeight
      margin-top: ($MAIN_HEADER_HEIGHT - $imgHeight) / 2
      margin-right: 5px

.functions
  margin: 0
  padding: 0
  list-style: none
  li
    float: left
    color: rgba($COLOR_WHITE, 0.5)
    margin: 0 15px
    &.active
      color: $COLOR_WHITE
.startedAt
  color: rgba($COLOR_WHITE, 0.5)
  margin-right: 15px
</style>

<script>
import {mapState} from 'vuex';
export default {
  data() {
    return {
      logo:
        'data:image/svg+xml;base64,Cgk8c3ZnIHN0eWxlPSJmaWxsOiM2MWRhZmIiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiIHZlcnNpb249IjEuMSIgIHZpZXdCb3g9IjAgMCAxMDAgMTAwIj4KCgoKCiAgICAgICAgICAgIDxnIHRyYW5zZm9ybT0iCiAgICAgICAgICAgICAgICAgICAgdHJhbnNsYXRlKDUwIDUwKQogICAgICAgICAgICAgICAgICAgIHRyYW5zbGF0ZSgwIDApICBzY2FsZSgxKSAgCiAgICAgICAgICAgICAgICAgICAgdHJhbnNsYXRlKC01MCAtNTApCiAgICAgICAgICAgICAgICAgICAgIj4KCgogICAgICAgICAgICAgICAgPGcgdHJhbnNmb3JtPSJzY2FsZSgzLjEwNTU5MDM1NjQ0NDM5ODcpIHRyYW5zbGF0ZSgwLjEwMDAwMDAyMzg0MTg1NzkxIDUuMjk5OTk4NDYyMjAwMTY1KSIgaWQ9InBpY3R1cmUiPjxwYXRoIGQ9Ik0yOCA5LjlhMSAxIDAgMSAxLTIgMCAxIDEgMCAxIDEgMiAweiBNMzAuOCA4LjRDMjkuMiA2LjMgMjYuNyA0LjcgMjQgNCAyMy40IDIuMSAyMS42LjggMTkuOC4zYy0yLS41LTQuMS0uMy01LjktLjEtLjQgMC0uNy4zLS44LjZzLS4xLjcuMiAxYy42LjcgMS4xIDEuNiAxLjUgMi41LTIuOC42LTUuMyAxLjktNy4zIDMuNkM2LjEgNi4yIDQuMSA1IDEuOSA0LjVjLS40LS4xLS45LjEtMS4xLjRTLjYgNS43LjggNmMyIDIuOSAxLjcgNy0uNiA5LjYtLjIuMy0uMy43LS4yIDFzLjUuNi44LjZjMi44LjMgNS42LTEuMiA2LjktMy41IDIuNCAxLjkgNS4zIDMuMiA4LjMgMy45LS4yLjktLjggMS43LTEuNiAyLjMtLjMuMi0uNS42LS40IDFzLjQuNy44LjhjLjQuMS44LjEgMS4yLjEgMS45IDAgMy43LS43IDUtMmwuOC0uOGMuNC0uNC43LS45IDEuMS0xLjEuNS0uMyAxLjEtLjQgMS44LS42LjMtLjEuNS0uMS44LS4yIDEuNy0uNCAzLjQtMS4zIDQuNy0yLjUgMS0uOSAxLjUtMS44IDEuNy0yLjcuMi0xLjEtLjItMi4zLTEuMS0zLjV6TTI5IDEzLjJjLTEuMSAxLTIuNCAxLjctMy44IDItLjIuMS0uNS4xLS43LjItLjguMi0xLjcuNC0yLjQuOS0uNy40LTEuMiAxLTEuNiAxLjUtLjIuMi0uNC41LS42LjctLjYuNi0xLjUgMS4xLTIuNCAxLjMuNS0uOS44LTEuOC44LTIuOCAwLS41LS4zLS45LS44LTEtMy40LS42LTYuNy0yLjItOS4zLTQuNS0uMi0uMi0uNC0uMy0uNy0uM2gtLjNjLS4zLjEtLjYuNC0uNy43LS41IDEuNS0xLjkgMi44LTMuNSAzLjIgMS4zLTIuNCAxLjUtNS4yLjctNy44IDEuMS42IDIuMSAxLjYgMi44IDIuNi4yLjMuNS41LjguNXMuNi0uMS44LS4zYzIuMS0yLjEgNS0zLjYgOC00LjEuMyAwIC41LS4yLjctLjQuMi0uMi4yLS41LjEtLjgtLjItMS0uNi0xLjktMS4xLTIuOCAxLjItLjEgMi4zLS4xIDMuNC4yIDEuMy4zIDIuOCAxLjMgMi45IDIuNyAwIC40LjQuOC44LjkgMi41LjUgNC44IDEuOSA2LjIgMy44LjYuOC45IDEuNS44IDIuMSAwIC40LS4zLjktLjkgMS41eiBNMjQuOCAxMi40Yy0uMy0uMS0uNS0uNS0uNi0uOS0uMS0uNS0uMS0xIC4xLTEuNC4xLS4yIDAtLjUtLjItLjctLjItLjEtLjUgMC0uNy4yLS4zLjYtLjQgMS40LS4yIDIuMS4yLjcuNyAxLjMgMS4yIDEuNmguMmMuMiAwIC40LS4xLjUtLjMgMC0uMi0uMS0uNS0uMy0uNnoiLz48L2c+CgoKICAgICAgICAgICAgPC9nPgoKCgk8L3N2Zz4K',
      functions: [
        {
          name: 'Directors',
          route: 'director',
        },
        {
          name: 'Performance',
          route: 'performance',
        },
        {
          name: 'Cached List',
          route: 'cachedList',
        },
      ],
    };
  },
  computed: {
    ...mapState({
      statsInfo: ({pike}) => pike.stats,
      currentRoute: ({route}) => route.name,
    }),
  },
};
</script>
