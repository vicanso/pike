import Vue from 'vue';
import _ from 'lodash';
import ElementUI from 'element-ui';
import 'element-ui/lib/theme-chalk/index.css';
import VuexRouterSync from 'vuex-router-sync';

import App from './app.vue';
import router from './router';
import store from './store';
import './request-interceptors';
import {getErrorMessage} from './helpers/util';
import {env} from './config';

Vue.use(ElementUI);

VuexRouterSync.sync(store, router);

// 注入 router 和 store
Vue.$router = router;
Vue.$store = store;

Vue.prototype.$loading = (options = {}) => {
  let loadingInstance = ElementUI.Loading.service(
    _.extend(
      {
        fullscreen: true,
        text: 'loading...',
      },
      options,
    ),
  );
  let resolved = false;
  const resolve = () => {
    if (resolved) {
      return;
    }
    resolved = true;
    loadingInstance.close();
  };
  setTimeout(resolve, options.timeout || 10 * 1000);
  return resolve;
};

Vue.prototype.$error = function $error(err) {
  const message = getErrorMessage(err);
  this.$message.error(message);
};

Vue.config.productionTip = env === 'production';

new Vue({
  router,
  store,
  render: h => h(App),
}).$mount('#app');
