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