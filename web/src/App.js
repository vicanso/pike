import React from 'react';
import { Route, HashRouter } from "react-router-dom";

import AppHeader from "./components/app_header";
import {
  CACHES_PATH,
  COMPRESSES_PATH,
  UPSTREAMS_PATH,
  LOCATIONS_PATH,
  SERVERS_PATH,
  ADMIN_PATH,
  HOME_PATH,
  INFLUXDB_PATH,
  CERTS_PATH,
  ALARMS_PATH
} from "./paths";
import Caches from "./components/caches";
import Compresses from "./components/compress";
import Upstreams from "./components/upstreams";
import Locations from "./components/locations";
import Servers from "./components/servers";
import Admin from "./components/admin";
import Home from "./components/home";
import Certs from "./components/certs";
import Influxdb from "./components/influxdb";
import Alarms from "./components/alarms";

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
          <Route path={ADMIN_PATH} component={Admin} />
          <Route path={CERTS_PATH} component={Certs} />
          <Route path={INFLUXDB_PATH} component={Influxdb} />
          <Route path={ALARMS_PATH} component={Alarms} />
          <Route path={HOME_PATH} exact component={Home} />
        </div>
      </HashRouter>
    </div>
  );
}

export default App;
