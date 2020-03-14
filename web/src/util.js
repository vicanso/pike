const hour = "h";
const minute = "m";
const second = "s"
const durations = {
  [second]: 10e9,
  [minute]: 60 * 10e9,
  [hour]: 60 * 60 * 10e9,
};

export function durationToNumber(d) {
  const unit = d.charAt(d.length - 1);
  return Number(d.substring(0, d.length - 1) * durations[unit]);
}

export function divideDuration(value) {
  const result = {
    unit: "",
    value: 0,
  };
  if (!value) {
    return result;
  }
  let done = false
  const units = [
    hour,
    minute,
    second,
  ];
  units.forEach((unit) => {
    if (done) {
      return;
    }
    const unitDuration = durations[unit];
    if(!unitDuration) {
      return;
    }
    if (value % unitDuration === 0) {
      result.unit = unit;
      result.value = (value / unitDuration);
      done = true;
    }
  });
  return result;
}

export function numberToDuration(value) {
  const result = divideDuration(value);
  if (!result.unit || !result.value) {
    return "";
  }
  return result.value  +result.unit;
}

export function toLocalTime(date) {
  const d = new Date(date);
  return `${d.toLocaleDateString()} ${d.toLocaleTimeString()}`;
}
