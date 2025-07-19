This package contains generated models and client for the immich api.

### Context
Immich has an API. It is under active development. The human-readable version of the API is [here](https://immich.app/docs/api/).

We want to work with the API based on its specs so would like client code for it.

The API is generated from open-api specs. The latest specs are [here](https://github.com/immich-app/immich/blob/main/open-api/immich-openapi-specs.json). To ensure compatibility with the server version log in to your immich instance and find the server version - this is likely in the bottom left side of the screen. The version allows you to specify the openapi specs file you want:

```
https://github.com/immich-app/immich/blob/v${your version number}$/open-api/immich-openapi-specs.json
```

The file can be downloaded to this directory.

### Generating the client
The generating library is [github.com/oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen). Follow its instruction on how to install.

The configuration file for our usecase is in [immich-openapi-config.yaml](./immich-openapi-config.yaml).

The immich openapi spec file as well as the config are referenced in [immich-generate.go](./immich-generate.go). Running the `go generate` in that file will produce the client file under [client.gen.go](./client.gen.go).

### Gotchas

1. The Immich API spec does not match what is used in practice. The API in practice are mounted under `${host}/api` but the spec does not explicitly include the `/api` part. As such all toolings must be modified in some way to work with the discrepancy.

2. The Immich API will respond differently when certain properties are set to null vs excluding it entirely. Make sure your request is not sending nulls when you did not want to send anything at all.
