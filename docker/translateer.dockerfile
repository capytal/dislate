FROM oven/bun:alpine AS build

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

RUN apk add --no-cache git
RUN git clone --depth=1 https://github.com/Songkeys/Translateer.git .

RUN bun install
RUN bun run build

FROM oven/bun:alpine AS run

COPY --from=build /usr/src/app/dist /usr/src/app
WORKDIR /usr/src/app

EXPOSE 8999

CMD ["bun", "run", "app.js"]
