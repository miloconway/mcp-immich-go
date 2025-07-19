# Model Context Protocol Server for Immich (Golang)

Work in progress.

Using unreleased (as of the time of writing) [MCP reference implementation in Go](https://github.com/modelcontextprotocol/go-sdk).

It also uses [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate client for [Immich's API](https://immich.app/docs/developer/open-api/).

# TBD

### No Real Auth
Currently capability is tested locally with a self-hosted Immich instance with a locally stored API Key

### Limited Integration
So far only implemented an MCP tool that executes a search against Immich and returns a few photos.
