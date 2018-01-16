import babel from 'rollup-plugin-babel';
import resolve from 'rollup-plugin-node-resolve';
import livereload from 'rollup-plugin-livereload';
import postcss from 'rollup-plugin-postcss';
import precss from 'precss';
import express from 'express';
import serveStatic from 'serve-static';
import httpProxy from 'http-proxy';

const proxy = httpProxy.createProxyServer({});

function serve() {
  const app = express()
  app.use(serveStatic('./dist'));
  app.use((req, res) => {
    proxy.web(req, res, {
      target: 'http://127.0.0.1:3015',
    });
  });
  app.listen(8080);
  return {
    name: 'serve',
    ongenerate: () => {
      console.info("the server is listen: http://127.0.0.1:8080/")
    },
  };
}

const plugins = [
  postcss({
    plugins: [
      precss(),
    ],
    extensions: ['.sss'],
    parser: 'sugarss',
  }),
  babel({
    babelrc: false,
    presets: ['es2015-rollup'],
    plugins: [['transform-react-jsx', {pragma: 'h'}]],
  }),
  resolve({
    jsnext: true,
  }),
  livereload(),
  serve(),
];

let config = {
  input: './src/app.js',
  output: {
    name: 'app',
    file: './dist/app.js',
    format: 'umd',
    sourcemap: true,
  },
  external: ['whatwg-fetch'],
  plugins: plugins,
};

export default config;