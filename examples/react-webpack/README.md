# @hmerritt/reactenv-webpack example

This example uses the `@hmerritt/reactenv-webpack` plugin for Webpack.

## Local setup

```sh
# install
yarn

# start locally
yarn start

# build
yarn build
```

---

Running the `reactenv` cli will inject the environment variables into your build.

> Every ENV value must be set. `reactenv` will error if even one is missing (you can set empty strings for optional values).

```sh
# run reactenv
reactenv run <path-to-asset-dir>
```

```sh
# run reactenv
reactenv run ./dist
```

## Usage

When running locally with `yarn start`, the plugin will passthrough and env variables so you can develop locally.

---

When building with `yarn build`, the plugin with replace all `process.env.*` values with a static string that can be replaced at a later time.

Before serving the build, you need to run the `reactenv` cli to inject the desired environment variables.

⚠️ It is crucial to note that until you run the `reactenv` cli, your built app will not have any environment variables ⚠️
