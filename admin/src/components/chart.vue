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
          datasets: [
            {
              label: '',
              backgroundColor: '#fff',
              borderColor: '#fff',
              data: [],
              fill: false,
            },
          ],
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
      _.forEach(data, item => {
        labels.push(dayjs(item.createdAt).format('hh:mm'));
        values.push(item[name] || 0);
      });
      configData.labels = labels;
      configData.datasets[0].data = values;
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
