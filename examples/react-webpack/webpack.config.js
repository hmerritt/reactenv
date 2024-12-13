const HtmlWebpackPlugin = require("html-webpack-plugin");
const path = require("path");
const dotenv = require("dotenv").config({ path: __dirname + "/.env" });

const ReactenvWebpackPlugin = require("@reactenv/webpack");
// const ReactenvWebpackPlugin = require("../../npm/plugin-webpack/lib");

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
