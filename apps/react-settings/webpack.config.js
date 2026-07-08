const path = require('path');

module.exports = {
  entry: { settings: './src/remote-entry.tsx' },
  output: {
    filename: '[name].js',
    libraryTarget: 'umd',
    library: ['ext-apps', '[name]'],
    publicPath: 'auto',
    path: path.resolve(__dirname, 'dist'),
    clean: true,
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader'],
      },
    ],
  },
  devServer: {
    port: 5174,
    hot: true,
    headers: {
      'Access-Control-Allow-Origin': '*',
    },
  },
};
