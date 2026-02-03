# VirtEngine TypeScript SDK examples

## How to run

Before running, the SDK needs to be built locally:

```sh
cd ts # navigate to ts folder, not examples
npm ci
npm run build
```

Afterwards use any preferred ts compilation command:

```sh
node --experimental-strip-types examples/create-deployment.ts # nodejs >=22
# or
bun run examples/create-deployment.ts
# or
deno run --allow-net --allow-env examples/create-deployment.ts
```

## React example

```sh
cd examples/react
npm install
npm run dev
```

## Vue example

```sh
cd examples/vue
npm install
npm run dev
```
