import React from "react";
import { Link, withRouter } from "react-router-dom";
import { Menu, Dropdown, Icon } from "antd";

import logo from "../../logo.svg";
import "./app_header.sass";
import i18n, { changeToEnglish, changeToChinese } from "../../i18n";
import {
  CACHES_PATH,
  COMPRESSES_PATH,
  UPSTREAMS_PATH,
  LOCATIONS_PATH,
  SERVERS_PATH,
  ADMIN_PATH
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
  },
  {
    name: i18n("nav.admin"),
    path: ADMIN_PATH
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
  renderLanguageSelector() {
    const menu = (
      <Menu>
        <Menu.Item>
          <a
            href="/en"
            onClick={e => {
              e.preventDefault();
              changeToEnglish();
              window.location.reload();
            }}
          >
            English
          </a>
        </Menu.Item>
        <Menu.Item>
          <a
            href="/zh"
            onClick={e => {
              e.preventDefault();
              changeToChinese();
              window.location.reload();
            }}
          >
            中文
          </a>
        </Menu.Item>
      </Menu>
    );
    return (
      <div className="langSelector">
        <Dropdown overlay={menu}>
          <span>
            {i18n("common.lang")} <Icon type="down" />
          </span>
        </Dropdown>
      </div>
    );
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
        {this.renderLanguageSelector()}
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
