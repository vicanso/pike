import React from "react";
import { Route } from "react-router-dom";

import AppHeader from "./components/app_header";
import Director from "./components/director";
import Caches from "./components/caches";
import Performance from "./components/performance";
import Config from "./components/config";
import AddUpstream from "./components/add_upstream";
import {
  DIRECTOR_PATH,
  CACHES_PATH,
  PERFORMANCE_PATH,
  CONFIG_PATH,
  ADD_UPSTREAM_PATH
} from "./paths";

class App extends React.Component {
  render() {
    return (
      <div className="App">
        <AppHeader />
        <div>
          <Route path={CACHES_PATH} component={Caches} />
          <Route path={PERFORMANCE_PATH} component={Performance} />
          <Route path={CONFIG_PATH} component={Config} />
          <Route path={ADD_UPSTREAM_PATH} component={AddUpstream} />
          <Route exact path={DIRECTOR_PATH} component={Director} />
        </div>
      </div>
    );
  }
}

export default App;
