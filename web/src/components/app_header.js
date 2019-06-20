import React from "react";
import { Link } from "react-router-dom";
import { withRouter } from "react-router-dom";

import logo from "../logo.svg";
import "./app_header.sass";
import {
  DIRECTOR_PATH,
  CACHES_PATH,
  PERFORMANCE_PATH,
  CONFIGS_PATH
} from "../paths";

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
    active: -1
  };
  render() {
    const { active } = this.state;
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
}

export default withRouter(props => <AppHeader {...props} />);
