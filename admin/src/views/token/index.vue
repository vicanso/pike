<template lang="pug">
.tokenPage
  el-form.form(
    ref='form'
    :model='form'
    label-width='60px'
  )
    el-form-item(
      label='Token'
    )
      el-input(
        v-model='form.token'
        placeholder="please input admin token"
        autofocus
      )
    el-form-item
      el-button(
        type='primary'
        @click='onSubmit'
        :style=`{
          width: '100%',
        }`
      ) Submit
</template>

<script>
import {
  saveAdminToken
} from '../../helpers/util';
export default {
  data() {
    return {
      form: {},
    };
  },
  methods: {
    onSubmit() {
      try {
        const {
          token,
        } = this.form;
        if (!token) {
          throw new Error('token can not be null')
        }
        saveAdminToken(token)
        this.$router.back();
        setTimeout(() => {
          location.reload();
        }, 30)
      } catch (err) {
        this.$error(err)
      }
    },
  },
}
</script>
<style lang="sass" scoped>
.tokenPage
  width: 600px
  margin: auto
  padding-top: 80px
  .form
    padding: 40px
    padding-bottom: 0
</style>


