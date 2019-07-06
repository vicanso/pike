import React from "react";
import { Link } from "react-router-dom";
import { withRouter } from "react-router-dom";
import request from "axios";
import { message } from "antd";

import logo from "../logo.svg";
import "./app_header.sass";
import {
  DIRECTOR_PATH,
  CACHES_PATH,
  PERFORMANCE_PATH,
  CONFIGS_PATH
} from "../paths";
import {
  CONFIGS,
} from "../urls";

const paths = [
  {
    name: "Directors",
    path: DIRECTOR_PATH
  },
  {
    name: "Caches",
    path: CACHES_PATH
  },
  {
    name: "Performance",
    path: PERFORMANCE_PATH
  },
  {
    name: "Configs",
    path: CONFIGS_PATH
  }
];

class AppHeader extends React.Component {
  state = {
    active: -1,
    version: '',
  };
  render() {
    const { active, version } = this.state;
    const arr = paths.map((item, index) => {
      let className = "";
      if (index === active) {
        className = "active";
      }
      return (
        <li key={item.name}>
          <Link to={item.path} className={className}>
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
          {version && <span className="version">
            {version}
          </span>}
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
  componentWillReceiveProps(newProps) {
    this.changeActive(newProps.location.pathname);
  }
  componentWillMount() {
    this.changeActive(this.props.location.pathname);
  }
  async componentDidMount() {
    try {
      const {
        data,
      } = await request.get(CONFIGS);
      this.setState({
        version: data.applicationInfo.version,
      });
    } catch (err) {
      message.error(err.message)
    }
  }
}

export default withRouter(props => <AppHeader {...props} />);
