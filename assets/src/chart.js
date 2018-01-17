import 'chart.js/dist/Chart.js'
import moment from 'moment';


Chart.defaults.global.defaultFontColor = '#fff';
const charts = {};

export function init(ctx, name) {
  if (charts[name]) {
    return charts[name];
  }
  const config = {
    type: 'line',
    data: {
      labels: [],
      datasets: [{
        label: '',
        backgroundColor: '#fff',
        borderColor: '#fff',
        data: [
        ],
        fill: false,
      }],
    },
    options: {
      gridLines: {
        display: false,
      },
      legend: {
        display: false,
      },
      responsive: true,
      maintainAspectRatio: false,
      title: {
        display: false,
      },
      tooltips: {
        mode: 'index',
        intersect: false,
      },
      hover: {
        mode: 'nearest',
        intersect: true,
      },
      scales: {
        xAxes: [{
          display: true,
        }],
        yAxes: [{
          display: true,
        }],
      },
    },
  };
  charts[name] = new Chart(ctx, config);
  return charts[name]; 
}

export function remove(name) {
  delete charts[name]
}

export function setData(name, data) {
  const chart = charts[name];
  if (!chart) {
    return;
  }
  const configData = chart.config.data;
  configData.labels = data.map(item => item.time);
  configData.datasets[0].data = data.map(item => item.value);
  chart.update()
}
