const ReactenvWebpackPlugin = require("./reactenv-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
var webpack = require("webpack");
const path = require("path");

const dotenv = require("dotenv").config({ path: __dirname + "/.env" });
const isDevelopment = process.env.NODE_ENV !== "production";

module.exports = {
    entry: "./src/index.js",
    output: {
        filename: "bundle.[fullhash].js",
        path: path.resolve(__dirname, "dist"),
        clean: true,
    },
    plugins: [
        new ReactenvWebpackPlugin(),
        new HtmlWebpackPlugin({
            template: "./src/index.html",
        }),
        new webpack.EnvironmentPlugin({ ...process.env }),
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
