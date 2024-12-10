# reactENV

Inject environment variables into a **built** react app (after `npm build`).

> Build once, configure later.

Useful for creating generic Docker images. Build your app once and add build files into Docker image, then configure at runtime without needing to install dependencies and build each time.

### Features ⚡

-   No runtime overhead
-   No app code changes required
-   Injection is strict by default, and will error if any values are missing
-   Blazing fast environment variable injection (~0.5ms for a basic react app)
-   (Optional) Bundler plugins to automate processing `process.env` values during build
    -   [Webpack plugin `@reactenv/webpack`](https://github.com/hmerritt/reactenv/tree/master/packages/plugin-webpack)

### Jump to:

-   [Install](#install)
-   [Usage](#usage)
-   [Example](#example)
-   [Reasoning](#reasoning)
-   [Aims](#aims)
-   [Licence](#licence)

## Install

Grab the latest binary from the releases page [here](https://github.com/hmerritt/reactenv/releases/latest).

Or install globally from npm:

```sh
npm i -g @reactenv/cli
```

Verify install by running `reactenv`, it should print the help:

```sh
reactenv
```

## Usage

### App

No code changes are required. You can use `process.env` to access environment variables as usual.

The magic happens at build-time. You have two options:

1. Manually set the value of every env variable to `__reactenv.<name>` at build (this option offers the most control, and is potentially more robust)

2. Use one of the bundler plugins to do it for you
    - [Webpack plugin `@reactenv/webpack`](https://github.com/hmerritt/reactenv/tree/master/packages/plugin-webpack)
    - (more coming soon)

### Injection via `reactenv`

After building your app, you should have a final bundle with all environment variables replaced with `__reactenv.<name>`.

`reactenv` is a CLI program used to replace all instances of `__reactenv.<name>` with actual values.

It uses the current host enviroment variables and will replace all matches in the bundle. (support for `.env` files is coming soon).

All you need to do is run `reactenv run <path-to-js-files>` and it will do it's thing:

```sh
# Inject environment variables into all `.js` files in `dist` directory
$ reactenv run dist
```

After running `reactenv`, your app is ready to be deployed and served!

---

Basic usage example:

```sh
# build app
$ npm run build

# Example file with un-replaced environment variables
$ cat dist/bundle.js
const apiUrl = "__reactenv.REACT_APP_API_URL";

# Set environment variable
$ REACT_APP_API_URL="https://api.example.com"

# Inject environment variables into all `.js` files in `dist` directory
$ reactenv run dist

$ cat dist/bundle.js
const apiUrl = "https://api.example.com";
```

## Example

For detailed examples, [go here](https://github.com/hmerritt/reactenv/tree/master/examples).

---

### Dockerfile example

```Dockerfile
# File: Dockerfile

# Build stage - install, build
FROM node as build
WORKDIR /app
COPY ./ /app/
ARG REACT_APP_NAME=__reactenv.REACT_APP_NAME
ARG REACT_APP_API_URL=__reactenv.REACT_APP_API_URL  # set all env values to be replaced
RUN npm install
RUN npm run build

# Final stage, production environment - use build, reactENV
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
RUN apk add --no-cache wget unzip libc6-compat
RUN wget https://github.com/hmerritt/reactenv/releases/download/v0.1.47/reactenv_0.1.47_linux_amd64.zip \
    && unzip reactenv_0.1.47_linux_amd64.zip \
    && chmod +x reactenv \
    && mv reactenv /usr/local/bin/ \
    && rm reactenv_0.1.47_linux_amd64.zip
ENTRYPOINT ["sh", "docker-entrypoint.sh"]
```

```sh
# File: docker-entrypoint.sh

reactenv /usr/share/nginx/html            # run reactenv in build directory

if [ "${?}" != "0" ]; then                # exit entrypoint script if reactenv failed
    exit 1
fi

nginx -g daemon off;
```

```sh
# File: docker-compose.yml

services:
  app:
    build:
      context: .
      dockerfile: ./Dockerfile

    ports:
      - "80:80"

    environment:
      - REACT_APP_NAME=My App
      - REACT_APP_API_URL=https://api.example.com

    restart: on-failure
```

## Reasoning

When creating a Docker image for a `React.js` app, there are few ways to change the environment:

1. Build react.js at container runtime (bad idea for many reasons)
2. Build specific Docker images for different environments (good for private images, but not for public ones with lots of configuration options)
3. Create an `env.js` file that contains environment variables and load it separately from HTML (better, but not ideal since it's adding to the total requests the end-user makes)

I wanted to create a fourth option, one that attempts to solve the problems of the other two solutions.

I'm aware that this solution has it's drawbacks and I don't recommend it for everyone. My hope is that as this program matures and becomes more robust, it could be relied upon and used without hesitation.

## Aims

Since this is being ran **after** a build, this program needs to be 100% reliable. If somthing does go wrong, it catches and reports it so a failed build does not end up in production.

-   Fast
-   Reliable
-   Easy to **debug**
-   Simple to use

## Licence

Apache-2.0 License
