# @reactenv/webpack

The package contains [`reactenv`](https://github.com/hmerritt/reactenv) plugin for Webpack.

> See [`reactenv`](https://github.com/hmerritt/reactenv) repo for more information.

## Installation

```sh
# npm
npm i -D @reactenv/webpack
# yarn
yarn add -D @reactenv/webpack
# pnpm
pnpm add -D @reactenv/webpack
```

## Usage

See [example webpack](https://github.com/hmerritt/reactenv/blob/master/examples/react-webpack/README.md) for an example project.

---

`webpack.config.js` import the webpack plugin `@reactenv/webpack`.

⚠️ This plugin cannot be used with `webpack.EnvironmentPlugin`. Remove it, and instead pass the same props into `@reactenv/webpack` as a replacement. ⚠️

```js
const ReactenvWebpackPlugin = require('@reactenv/webpack');

module.exports = {
    plugins: [new ReactenvWebpackPlugin({ ...process.env, ...dotenv.parsed })],
};
```
