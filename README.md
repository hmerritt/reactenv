# reactENV

Inject `REACT_APP_` environment variables into a **built** react app (after `npm build`).

Used as an abstraction layer to build react once, then configure later.

Useful for adding ENV config to Docker images, **without** having to build react every time you start the container.

## Usage

Example Dockerfile build

```Dockerfile
# Build stage - copy, install, build, reactENV
FROM node as build-stage
WORKDIR /app
COPY ./ /app/

RUN apk add --no-cache libc6-compat  # needed for Go programs to run
RUN npm install
RUN npm run build

RUN wget URL -o reactenv  # download reactENV
RUN reactenv build        # run reactenv in build directory
```

## Reasoning

When creating a Docker image for a `React.js` app, there are few ways to change the environment:

1. Build react.js at container runtime (bad idea for many reasons)
2. Build specific Docker images for different environments (good for private images, but not for public ones with lots of configuration options)
3. Create an `env.js` file that contains environment variables and load it separately from HTML (better, but not ideal since it's adding to the total requests the end-user makes)

I wanted to create a fourth option, one that attempts to solve the problems of the other two solutions.

I'm aware that this solution has it's drawbacks and I don't recommend it for everyone. My hope is that as this program matures and becomes more robust, it could be relied upon and used without hesitation.

If there is a better way that I don't know about **please** create an issue and tell me!

## Aims

Since this is being ran **after** a build, this program needs to be 100% reliable. If somthing does go wrong, it catches and reports it so a failed build does not end up in production.

-   Reliable
-   Easy to **debug**
-   Simple to use

## Licence

Apache-2.0 License
