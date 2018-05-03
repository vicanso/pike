<template lang="pug">
  #app
    main-header
    transition
      .contentWrapper: router-view
</template>

<script>
import {mapActions} from 'vuex';
import _ from 'lodash';

import MainHeader from './components/main-header';
export default {
  components: {
    MainHeader,
  },
  methods: {
    ...mapActions(['getStats']),
  },
  async beforeMount() {
    const close = this.$loading();
    try {
      await this.getStats();
      this.$router.push('director');
    } catch (err) {
      if (_.get(err, 'response.status') === 401) {
        this.$router.push('token');
        return;
      }
      this.$error(err);
    } finally {
      close();
    }
  },
};
</script>

<style src="./app.sass" lang="sass"></style>
