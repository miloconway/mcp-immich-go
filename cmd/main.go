// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/miloconway/mcp-immich/immichapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout")

type SearchArgs struct {
	Query string `json:"query" mcp:"the query used to find photos"`
}

func makeSearch() mcp.ToolHandlerFor[SearchArgs, struct{}] {
	c := &immichapi.Client{
		Server: os.Getenv("SERVER_HOST"),
		Client: &http.Client{},
		RequestEditors: []immichapi.RequestEditorFn{
			func(ctx context.Context, req *http.Request) error {
				// every request must have an API key for validation, either as a query param or a header
				req.Header.Add("x-api-key", os.Getenv("IMMICH_API_KEY"))
				// the paths in the spec does not match what is used in practice, this aligns them
				req.URL.Path = path.Join("/api", req.URL.Path)
				return nil
			},
		},
	}

	return func(
		ctx context.Context,
		ss *mcp.ServerSession,
		params *mcp.CallToolParamsFor[SearchArgs],
	) (*mcp.CallToolResultFor[struct{}], error) {

		// do a smart search
		var page float32 = 1.0
		var size float32 = 2.0
		var assetType = immichapi.IMAGE
		requestBody := immichapi.SearchSmartJSONRequestBody{
			Query: params.Arguments.Query,
			Page:  &page,
			Size:  &size,
			Type:  &assetType,
		}

		resp, err := c.SearchSmart(ctx, requestBody)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			dump, _ := httputil.DumpRequest(resp.Request, true)
			return nil, fmt.Errorf("no smart search response: %s", dump)
		}

		parsedResults, err := immichapi.ParseSearchSmartResponse(resp)

		if err != nil {
			return nil, err
		}

		thumbnails := make([]mcp.Content, len(parsedResults.JSON200.Assets.Items))

		// TODO: make this concurrent
		for i, item := range parsedResults.JSON200.Assets.Items {
			id, err := uuid.Parse(item.Id)

			if err != nil {
				return nil, err
			}

			downloadResp, err := c.DownloadAsset(ctx, id, nil)

			if err != nil {
				return nil, err
			}

			downloadResults, err := immichapi.ParseDownloadAssetResponse(downloadResp)

			if err != nil {
				return nil, err
			}

			thumbnails[i] = &mcp.ImageContent{
				Data:     downloadResults.Body,
				MIMEType: *item.OriginalMimeType,
			}
		}

		return &mcp.CallToolResultFor[struct{}]{
			Content: thumbnails,
		}, nil
	}
}

func PromptSearch(
	ctx context.Context,
	ss *mcp.ServerSession,
	params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {

	return &mcp.GetPromptResult{
		Description: "Photo search prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: "Search for person, place, or thing",
				},
			},
		},
	}, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{
		Name: "immich",
	}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name: "search",
		Description: `
Find my photos using a query. The query will work best with locations - 
such as 'San Francisco', or a person's name - such as 'Mitchell'. Descriptions of 
the photo such as 'forest' may also work.`,
	}, makeSearch())
	server.AddPrompt(&mcp.Prompt{
		Name: "query",
	}, PromptSearch)
	server.AddResource(&mcp.Resource{
		Name:     "info",
		MIMEType: "text/plain",
		URI:      "embedded:info",
	}, handleEmbeddedResource)

	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		log.Printf("MCP handler listening at %s", *httpAddr)
		http.ListenAndServe(*httpAddr, handler)
	} else {
		t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
		if err := server.Run(context.Background(), t); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}
}

var embeddedResources = map[string]string{
	"info": "This is the immich search server.",
}

func handleEmbeddedResource(_ context.Context, _ *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	u, err := url.Parse(params.
		URI)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "embedded" {
		return nil, fmt.Errorf("wrong scheme: %q", u.Scheme)
	}
	key := u.Opaque
	text, ok := embeddedResources[key]
	if !ok {
		return nil, fmt.Errorf("no embedded resource named %q", key)
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: params.URI, MIMEType: "text/plain", Text: text},
		},
	}, nil
}
