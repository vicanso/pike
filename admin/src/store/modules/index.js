import * as pikeModule from './pike';

const modules = {};
const actions = {};
const getters = {};


modules.pike = pikeModule.default;
Object.assign(actions, pikeModule.actions);

export default {
  actions,
  modules,
  getters,
};