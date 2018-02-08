import babel from 'rollup-plugin-babel';
import resolve from 'rollup-plugin-node-resolve';
import uglify from 'rollup-plugin-uglify';
import postcss from 'rollup-plugin-postcss';
import precss from 'precss';
import rimraf from 'rimraf';

rimraf.sync('./dist/*.map')

const plugins = [
  postcss({
    plugins: [
      precss(),
    ],
    extensions: ['.sss'],
    parser: 'sugarss',
  }),
  resolve({
    jsnext: true,
    main: true
  }),
  babel({
    babelrc: false,
    presets: ['es2015-rollup'],
    plugins: [['transform-react-jsx', { pragma: 'h' }]],
  }),
  resolve({
    jsnext: true,
  }),
  uglify()
]

let config = {
  input: './src/app.js',
  output: {
    name: 'app',
    file: './dist/app.js',
    format: 'umd',
    sourcemap: false, 
  },
  external: ['whatwg-fetch'],
  plugins: plugins
}

export default config