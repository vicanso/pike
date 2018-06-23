
import {mapActions, mapState} from 'vuex';
import _ from 'lodash';

export default {
  data() {
    return {
      keyword: '',
    };
  },
  methods: {
    ...mapActions(['getFetching']),
  },
  computed: {
    ...mapState({
      fetchings: ({pike}) => pike.fetchings,
    }),
    fetchingList: function() {
      const {
        fetchings,
        keyword,
      } = this;
      const reg = new RegExp(keyword, 'gi');
      return _.filter(fetchings, (item) => {
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
      await this.getFetching()
    } catch (err) {
      this.$error(err)
    } finally {
      close();
    }
  },
}