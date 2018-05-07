
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