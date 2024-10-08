const path = require("path");
const HtmlWebpackPlugin = require('html-webpack-plugin');
const {CleanWebpackPlugin} = require("clean-webpack-plugin");
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = {
  entry: {
    index: "./src/index.ts",
    recorder: "./src/recorder.ts",
  },
  mode: "production",
  devtool: 'source-map',
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: "ts-loader",
        exclude: /node_modules/,
      },
      {
        test: /\.css$/i,
        use: [
          MiniCssExtractPlugin.loader,
          'css-loader'
        ],
      },
    ],
  },
  resolve: {
    extensions: [".tsx", ".ts", ".js"],
  },
  output: {
    filename: '[name].[chunkhash].js', // <- ensure unique bundle name
    path: path.resolve(__dirname, 'dist'),
    publicPath: "/spine/static/",
  },
  optimization: {
    splitChunks: {
      chunks: 'all',
    },
  },
  plugins: [
    new CleanWebpackPlugin(),
    new HtmlWebpackPlugin({
      filename: "index.html",
      chunks: ["index"],
      template: path.resolve(__dirname, "./src/templates/index.html")
    }),
    new HtmlWebpackPlugin({
      filename: "recorder.html",
      chunks: ["recorder"],
      template: path.resolve(__dirname, "./src/templates/recorder.html")
    }),
    new MiniCssExtractPlugin({filename: "styles.[chunkhash].css"})
  ]
};

