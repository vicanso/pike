<template lang="pug">
.chart
  canvas(
    ref="canvas"
  )
</template>
<script>
import Chart from 'chart.js/dist/Chart.js';
import _ from 'lodash';
import dayjs from 'dayjs';

Chart.defaults.global.defaultFontColor = '#fff';
const colors = [
  'rgba(255, 255, 255, 1)',
  'rgba(255,99,132,1)',
  'rgba(54, 162, 235, 1)',
  'rgba(255, 206, 86, 1)',
  'rgba(75, 192, 192, 1)',
  'rgba(153, 102, 255, 1)',
  'rgba(255, 159, 64, 1)',
];

export default {
  name: 'performance-chart',
  props: ['data', 'name'],
  data() {
    return {};
  },
  methods: {
    initChart() {
      if (this.chart) {
        return this.chart;
      }
      const {canvas} = this.$refs;
      const config = {
        type: 'line',
        data: {
          labels: [],
          datasets: [],
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
            xAxes: [
              {
                display: true,
              },
            ],
            yAxes: [
              {
                display: true,
              },
            ],
          },
        },
      };
      this.chart = new Chart(canvas, config);
      return this.chart;
    },
    fresh() {
      const chart = this.initChart();
      const {name, data} = this;
      const configData = chart.config.data;
      const labels = [];
      const values = [];
      const getDefaultValue = (label, colorIndex = 0) => ({
        backgroundColor: colors[colorIndex],
        borderColor: colors[colorIndex],
        fill: false,
        label,
        data: [],
      });
      _.forEach(data, item => {
        labels.push(dayjs(item.createdAt).format('hh:mm'));
        let value = item[name] || 0;
        if (_.isNumber(value)) {
          if (!values[0]) {
            values[0] = getDefaultValue(name, 0);
          }
          values[0].data.push(value);
          return;
        }
        let index = 0;
        _.forEach(value, (v, k) => {
          if (!values[index]) {
            values[index] = getDefaultValue(`${name}-${k}`, index);
          }
          values[index].data.push(v);
          index += 1;
        });
      });
      configData.labels = labels;
      configData.datasets = values;
      chart.update();
    },
  },
  mounted() {
    this.fresh();
  },
  watch: {
    data() {
      this.fresh();
    },
  },
};
</script>
<style lang="sass" scoped>
.chart, canvas
  height: 100%
  width: 100%
</style>
