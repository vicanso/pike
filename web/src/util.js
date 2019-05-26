export function getURLPrefix() {
  let prefix = "";
  const { pathname } = window.location;
  if (pathname !== "/") {
    const arr = pathname.split("/");
    prefix = `/${arr[1]}`;
  }
  return prefix;
}

const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;

export function getExpiredDesc(seconds) {
  if (seconds >= day) {
    return Math.ceil(seconds / day, 2) + " D";
  }
  if (seconds >= hour) {
    return Math.ceil(seconds / hour, 2) + " h";
  }
  if (seconds >= minute) {
    return Math.ceil(seconds / minute, 2) + " m";
  }
  return `${seconds} s`;
}
