import Vue from 'vue';
import Vuex from 'vuex';
import createLogger from 'vuex/dist/logger';

import modules from './modules';
import {env} from '../config';

Vue.use(Vuex);

const isProduction = env === 'production';
const debug = !isProduction;

const store = new Vuex.Store({
  ...modules,
  strict: debug,
  // store中使用的插件
  plugins: debug ? [createLogger()] : [],
});

export default store;
