const {
  getHTMLPlugins,
  getOutput,
  getCopyPlugins,
  getFirefoxCopyPlugins,
  getEntry,
  getDefinePlugin,
  getChromeEntry
} = require('./webpack.utils');
const config = require('./config.json');

const generalConfig = {
  mode: 'development',
  devtool: 'source-map',
  module: {
    rules: [
      {
        loader: 'babel-loader',
        exclude: /node_modules/,
        test: /\.(js|jsx)$/,
        query: {
          presets: ['@babel/preset-env', '@babel/preset-react']
        },
        resolve: {
          extensions: ['.js', '.jsx']
        }
      },
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['eslint-loader']
      },
      {
        test: /\.scss$/,
        use: [
          {
            loader: 'style-loader'
          },
          {
            loader: 'css-loader'
          },
          {
            loader: 'sass-loader'
          }
        ]
      }
    ]
  }
};

module.exports = [
  {
    ...generalConfig,
    entry: getChromeEntry(config.chromePath),
    output: getOutput('chrome', config.devDirectory),
    plugins: [
      ...getHTMLPlugins('chrome', config.devDirectory, config.chromePath),
      ...getCopyPlugins('chrome', config.devDirectory, config.chromePath),
      ...getDefinePlugin(JSON.stringify('http://localhost:3001/'))
    ]
  },
  {
    ...generalConfig,
    entry: getEntry(config.operaPath),
    output: getOutput('opera', config.devDirectory),
    plugins: [
      ...getHTMLPlugins('opera', config.devDirectory, config.operaPath),
      ...getCopyPlugins('opera', config.devDirectory, config.operaPath)
      // new webpack.DefinePlugin({
      //   'process.env.API_URL': 'http://localhost:3001/'
      // })
    ]
  },
  {
    ...generalConfig,
    entry: getEntry(config.firefoxPath),
    output: getOutput('firefox', config.devDirectory),
    plugins: [
      ...getFirefoxCopyPlugins(
        'firefox',
        config.devDirectory,
        config.firefoxPath
      ),
      ...getHTMLPlugins('firefox', config.devDirectory, config.firefoxPath),
      ...getDefinePlugin(JSON.stringify('http://localhost:3001/'))
    ]
  }
];
