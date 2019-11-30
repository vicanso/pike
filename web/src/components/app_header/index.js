import React from "react";
import { Link, withRouter } from "react-router-dom";

import logo from "../../logo.svg";
import "./app_header.sass";
import i18n from "../../i18n";
import {
  CACHES_PATH,
  COMPRESSES_PATH,
  UPSTREAMS_PATH,
  LOCATIONS_PATH,
  SERVERS_PATH
} from "../../paths";

const paths = [
  {
    name: i18n("nav.caches"),
    path: CACHES_PATH
  },
  {
    name: i18n("nav.compresses"),
    path: COMPRESSES_PATH
  },
  {
    name: i18n("nav.upstreams"),
    path: UPSTREAMS_PATH
  },
  {
    name: i18n("nav.locations"),
    path: LOCATIONS_PATH
  },
  {
    name: i18n("nav.servers"),
    path: SERVERS_PATH
  }
  // {
  //   name: "Performance"
  //   // path: PERFORMANCE_PATH
  // },
  // {
  //   name: "Configs"
  //   // path: CONFIGS_PATH
  // }
];

class AppHeader extends React.Component {
  state = {
    active: -1,
    version: ""
  };
  componentDidMount() {
    const { location } = this.props;
    this.changeActive(location.pathname);
  }
  render() {
    const { active, version } = this.state;
    const arr = paths.map((item, index) => {
      let className = "";
      if (index === active) {
        className = "active";
      }
      return (
        <li key={item.name}>
          <Link
            to={item.path}
            className={className}
            onClick={() => {
              this.setState({
                active: index
              });
            }}
          >
            {item.name}
          </Link>
        </li>
      );
    });
    return (
      <div className="AppHeader clearfix">
        <div className="logo">
          <img src={logo} alt="logo" />
          Pike
          {version && <span className="version">{version}</span>}
        </div>
        <ul className="functions">{arr}</ul>
      </div>
    );
  }
  changeActive(routePath) {
    let active = -1;
    paths.forEach((item, index) => {
      if (item.path === routePath) {
        active = index;
      }
    });
    this.setState({
      active
    });
  }
}

export default withRouter(props => <AppHeader {...props} />);
