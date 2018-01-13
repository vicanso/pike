import babel from 'rollup-plugin-babel';
import resolve from 'rollup-plugin-node-resolve';
import livereload from 'rollup-plugin-livereload';
import postcss from 'rollup-plugin-postcss';
import serve from 'rollup-plugin-serve';
import precss from 'precss';

const plugins = [
  postcss({
    plugins: [
      precss(),
    ],
    // extensions: ['.sss'],
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
  serve({
    contentBase: './dist/',
    port: 8080,
    // open: true,
  }),
];

let config = {
  input: './src/app.js',
  output: {
    name: 'app',
    file: './dist/app.js',
    format: 'umd',
    sourcemap: true,
  },
  plugins: plugins,
};

export default config;