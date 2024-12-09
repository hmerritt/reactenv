const ReactenvWebpackPlugin = require("../../packages/plugin-webpack/lib");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const path = require("path");
const dotenv = require("dotenv").config({ path: __dirname + "/.env" });

const webpack = require("webpack");

module.exports = {
	entry: "./src/index.js",
	mode: "development",
	output: {
		filename: "bundle.[fullhash].js",
		path: path.resolve(__dirname, "dist"),
		clean: true,
	},
	plugins: [
		new ReactenvWebpackPlugin({ ...process.env, ...dotenv.parsed }),
		new HtmlWebpackPlugin({
			template: "./src/index.html",
		}),
		new webpack.EnvironmentPlugin
	],
	resolve: {
		modules: [__dirname, "src", "node_modules"],
		extensions: [".js", ".jsx", ".tsx", ".ts"],
	},
	module: {
		rules: [
			{
				test: /\.jsx?$/,
				exclude: /node_modules/,
				use: ["babel-loader"],
			},
			{
				test: /\.css$/,
				exclude: /node_modules/,
				use: ["style-loader", "css-loader"],
			},
			{
				test: /\.(png|svg|jpg|gif)$/,
				exclude: /node_modules/,
				use: ["file-loader"],
			},
		],
	},
};
