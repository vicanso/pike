import React from 'react';
import { Route, HashRouter } from "react-router-dom";

import AppHeader from "./components/app_header";
import {
  CACHES_PATH,
  COMPRESSES_PATH,
  UPSTREAMS_PATH,
  LOCATIONS_PATH,
  SERVERS_PATH,
} from "./paths";
import Caches from "./components/caches";
import Compresses from "./components/compress";
import Upstreams from "./components/upstreams";
import Locations from "./components/locations";
import Servers from "./components/servers";

function App() {
  return (
    <div className="App">
      <HashRouter>
        <AppHeader />
        <div>
          <Route path={CACHES_PATH} component={Caches} />
          <Route path={COMPRESSES_PATH} component={Compresses} />
          <Route path={UPSTREAMS_PATH} component={Upstreams} />
          <Route path={LOCATIONS_PATH} component={Locations} />
          <Route path={SERVERS_PATH} component={Servers} />
        </div>
      </HashRouter>
    </div>
  );
}

export default App;
