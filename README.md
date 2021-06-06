# reactENV

Inject `REACT_APP_` environment variables into a **built** react app (after `npm build`).

Used as an abstraction layer to build react once, then configure later.

Useful for adding ENV config to Docker images, **without** having to build react every time you start the container.

## Usage

Example Dockerfile build

```Dockerfile
#
# File: Dockerfile
#

# Build stage - install, build, reactENV
FROM node as build
WORKDIR /app
COPY ./ /app/
RUN npm install
RUN npm run build

# Final stage, production environment - use build, reactENV
FROM nginx:1.16.0-alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80

RUN apk add --no-cache wget libc6-compat       # needed for Go programs to run
RUN wget URL -o reactenv && chmod +x reactenv  # download reactENV + make it an executable

ENTRYPOINT ["sh", "docker-entrypoint.sh"]


#
# File: docker-entrypoint.sh
#

reactenv /usr/share/nginx/html            # run reactenv in build directory

if [ "${?}" != "0" ]; then                # exit entrypoint script if reactenv failed
    exit 1
fi

nginx -g daemon off;
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
