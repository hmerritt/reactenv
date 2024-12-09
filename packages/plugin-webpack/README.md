# @hmerritt/reactenv-webpack

The package contains reactenv plugin for Webpack.

## Installation

```sh
# npm
npm i -D @hmerritt/reactenv-webpack
# yarn
yarn add -D @hmerritt/reactenv-webpack
# pnpm
pnpm add -D @hmerritt/reactenv-webpack
```

## Usage

See [example webpack](../../examples/react-webpack/README.md) for an example project.

---

`webpack.config.js` import the webpack plugin `@hmerritt/reactenv-webpack`.

⚠️ This plugin cannot be used with `webpack.EnvironmentPlugin`. Remove it, and instead pass the same props into `@hmerritt/reactenv-webpack` as a replacement. ⚠️

```js
const ReactenvWebpackPlugin = require('@hmerritt/reactenv-webpack');

module.exports = {
    plugins: [new ReactenvWebpackPlugin({ ...process.env, ...dotenv.parsed })],
};
```
