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
          display: false,
        }],
        yAxes: [{
          display: false,
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

export function addData(name, value) {
  const chart = charts[name];
  if (!chart) {
    return;
  }
  const data = chart.config.data;
  data.labels.push(moment().format('HH:mm'))
  data.datasets[0].data.push(value);
  chart.update()
  // config.data.labels.push(month);

  //               config.data.datasets.forEach(function(dataset) {
  //                   dataset.data.push(randomScalingFactor());
  //               });
}
