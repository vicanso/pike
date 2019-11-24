import React from 'react';
import { Route, HashRouter } from "react-router-dom";

import AppHeader from "./components/app_header";
import {
  CACHES_PATH,
  COMPRESSES_PATH,
} from "./paths";
import Caches from "./components/caches";
import Compresses from "./components/compress";

function App() {
  return (
    <div className="App">
      <HashRouter>
        <AppHeader />
        <div>
          <Route path={CACHES_PATH} component={Caches} />
          <Route path={COMPRESSES_PATH} component={Compresses} />
        </div>
      </HashRouter>
    </div>
  );
}

export default App;
