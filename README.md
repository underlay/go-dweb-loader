# go-dweb-loader

Sometimes you want to be sure that the remote JSON-LD contexts you're loading are really the ones the author intended, and a good way of doing this is content-addressed contexts using `ipfs://` or `dweb:/ipfs/` URI schemes.

`DwebDocumentLoader` uses the CoreAPI interface from `github.com/ipfs/interface-go-ipfs-core`, which is what you get passed if you write a [`go-ipfs` plugin](https://github.com/ipfs/go-ipfs/blob/master/docs/plugins.md) or if you connect to a remote IPFS node over HTTP with the [`go-ipfs-http-client`](https://github.com/ipfs/go-ipfs-http-client).

```golang
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	files "github.com/ipfs/go-ipfs-files"
	ipfs "github.com/ipfs/go-ipfs-http-client"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	ld "github.com/piprate/json-gold/ld"
)

func main() {
	api, err := ipfs.NewURLApiWithClient("http://localhost:5001", http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse("https://w3id.org/security/v1")
	if err != nil {
		log.Fatal(err)
	}

	resolved, err := api.Unixfs().Add(
		context.Background(),
		files.NewWebFile(u),
		options.Unixfs.CidVersion(1),
		options.Unixfs.RawLeaves(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	var doc map[string]interface{}
	err = json.Unmarshal([]byte(fmt.Sprintf(`{
		"@context": [
			"ipfs://%s",
			{ "ex": "http://example.org/vocab#" }
		],
		"@graph": {
			"ex:request": "DELETE /private/2840-credit-card-log"
		},
		"signature": {
			"@type": "GraphSignature2012",
			"creator": "http://example.com/people/john-doe#key-5",
			"nonce": "8495723045.84957",
			"signatureValue": "Q3ODIyOGQzNGVkMzVm4NTIyZ43OWM32NjITkZDYMmMzQzNmExMgoYzI="
		}
	}`, resolved.Cid().String())), &doc)
	if err != nil {
		log.Fatal(err)
	}

	proc := ld.NewJsonLdProcessor()
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = NewDwebDocumentLoader(api)
	expanded, err := proc.Expand(doc, opts)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(expanded, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
```

... which outputs the correct, expanded document

```json
[
	{
		"@graph": [
			{
				"http://example.org/vocab#request": [
					{
						"@value": "DELETE /private/2840-credit-card-log"
					}
				]
			}
		],
		"https://w3id.org/security#signature": [
			{
				"@type": ["https://w3id.org/security#GraphSignature2012"],
				"http://purl.org/dc/terms/creator": [
					{
						"@id": "http://example.com/people/john-doe#key-5"
					}
				],
				"https://w3id.org/security#nonce": [
					{
						"@value": "8495723045.84957"
					}
				],
				"https://w3id.org/security#signatureValue": [
					{
						"@value": "Q3ODIyOGQzNGVkMzVm4NTIyZ43OWM32NjITkZDYMmMzQzNmExMgoYzI="
					}
				]
			}
		]
	}
]
```

Usage should be fairly straightforward. `dag-cbor` IPLD formats are supported with `dweb:/ipld/` URIs if you want to be really compact.
