const path = require('path');
const webpackMerge = require('webpack-merge');
const configs = require('./config.js');
const common = require('./webpack.common.js');

const PUBLIC_PATH = '/assets/';
const isDev = process.env.NODE_ENV === 'development';

const base = {
  target: 'web',
  entry: {
    app: [
      'babel-polyfill',
      'script-loader!jquery',
      'script-loader!foundation-sites',
      path.join(__dirname, '../src/index.jsx'),
    ],
    vendor: [
      'babel-polyfill',
      'script-loader!jquery',
      'script-loader!foundation-sites',
    ],
  },
  output: {
    path: path.join(__dirname, '..', 'dist/server/public/assets'),
    filename: '[name].[chunkhash].js',
    chunkFilename: '[name].[chunkhash].js',
    publicPath: PUBLIC_PATH,
  },
};

const configureDevServer = () => {
  if (!isDev) {
    return {};
  }

  return {
    devServer: {
      contentBase: false,
      hot: false,
      inline: true,
      port: '8080',
      proxy: {
        '/': {
          target: 'http://localhost:4001',
        },
      },
      watchOptions: {
        aggregateTimeout: 300,
        poll: 1000,
      },
    },
  };
};

const config = webpackMerge(
  common,
  configs.configureStylesheet(),
  configs.configureAssets(),
  configs.configureVendorSplit(),
  configs.configureHTMLTemplate(),
  configs.configureEnv('browser'),
  configureDevServer(),
  base
);

module.exports = config;
