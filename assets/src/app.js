import { h, app } from 'hyperapp';
import 'whatwg-fetch';

import './global.sss';
import './app.sss';
import Performance from './performance';
import Directors from './directors';
import {
  state,
  actions,
  getStats,
  getDirectors,
} from './actions';


let refreshStatsInterval = null;
const refreshStats = () => {
  getStats().then((data) => {
    main.setStats(data);
  });
}

const view = (state, actions) => {
  const currentView = state.view;
  const getNav = (view, name) => {
    return <li><a
      href="javascript:;"
      class={currentView == view ? "active": ""}
      onclick={() => {
        clearInterval(refreshStatsInterval);
        if (view === 'performance') {
          refreshStatsInterval = setInterval(refreshStats, 60 * 1000);
          refreshStats();
        } else {
          actions.resetDirectors();
          getDirectors().then((data) => {
            main.setDirectors(data);
          });
        }
        actions.changeView(view);
      }}
    >
      {name} 
    </a></li> 
  }
  return <div>
    <nav class="navBar">Pike Dashboard
      <ul>
        {getNav("default", "Directors")}
        {getNav("performance", "Performance")}
      </ul>
      {
        state.uptime &&
        <div
          class="launthedAt gray"
          title={state.launchedAt}
        >
          launthed at:
          <span class="mleft5">{state.uptime}</span>
        </div>
      }
    </nav>
    {
      currentView == 'performance' && <Performance state={state} />
    }
    {
      currentView == 'default' && <Directors state={state} />
    }
  </div>
};

const main = app(state, actions, view, document.body)

getStats().then((data) => {
  main.setLaunchedAt(data.startedAt);
});

getDirectors().then((data) => {
  main.setDirectors(data);
});
