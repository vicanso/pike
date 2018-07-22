<template lang="pug">
  #app
    main-header(
      @togglePing="changePingStatus"
    )
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
    ...mapActions(['getStats', 'getPingStatus', 'togglePing']),
    intervalGetStats() {
      setInterval(() => {
        this.getStats();
      }, 60 * 1000);
    },
    async changePingStatus() {
      const close = this.$loading();
      try {
        await this.togglePing();
      } catch (err) {
        this.$error(err);
      } finally {
        close();
      }
    },
  },
  async beforeMount() {
    const close = this.$loading();
    try {
      await this.getStats();
      await this.getPingStatus();
      this.intervalGetStats();
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
