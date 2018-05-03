import Vue from 'vue';
import Router from 'vue-router';

import routes from './routes';

Vue.use(Router);

const router = new Router({
  mode: 'hash',
  routes,
});

let pageLoadStats = null;
router.beforeEach((to, from, next) => {
  pageLoadStats = {
    name: to.name,
    to: to.path,
    from: from.path,
    startedAt: Date.now(),
  };
  next();
});

router.afterEach(to => {
  if (!pageLoadStats || pageLoadStats.name !== to.name) {
    return;
  }
  const use = Date.now() - pageLoadStats.startedAt;
  pageLoadStats.use = use;
  // console.info(pageLoadStats);
  pageLoadStats = null;
});

export default router;
