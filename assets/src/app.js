import "bootstrap/dist/css/bootstrap.min.css"
import { h, app } from 'hyperapp'

// import './global.sss'
// import './app.sss'

const state = {
  count: 0
}

const actions = {
  down: () => state => ({ count: state.count - 1 }),
  up: () => state => ({ count: state.count + 1 })
}

const Pluzze = ({name, count}) => {
  const cls = 'green col-3';
  return <li class={cls}>
    {name}
  </li>
};


const view = (state, actions) => (
  <div>
    <nav class="navBar">Pike Dashboard</nav>
    <div class="contentWrapper container">
      <div class="bkz">
        <h3 class="bla blb">QUICK PERFORMANCE</h3>
      </div>
      <ul class="row">
        <Pluzze name={"mytest"} count={100} />
      </ul>
    </div>
  </div>
)

const main = app(state, actions, view, document.body)