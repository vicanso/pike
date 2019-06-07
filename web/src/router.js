import { createBrowserHistory } from "history";

const history = createBrowserHistory();

export function getHistory() {
  return history;
}

export function back() {
  history.goBack();
}
