package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"

	files "github.com/ipfs/go-ipfs-files"
	ipfs "github.com/ipfs/go-ipfs-http-client"
	options "github.com/ipfs/interface-go-ipfs-core/options"
	ld "github.com/piprate/json-gold/ld"
	loader "github.com/underlay/go-dweb-loader"
)

func TestSecurityExpansion(t *testing.T) {
	api, err := ipfs.NewURLApiWithClient("http://localhost:5001", http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://w3id.org/security/v1")
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := api.Unixfs().Add(
		context.Background(),
		files.NewWebFile(u),
		options.Unixfs.CidVersion(1),
		options.Unixfs.RawLeaves(true),
	)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	proc := ld.NewJsonLdProcessor()
	opts := ld.NewJsonLdOptions("")
	opts.DocumentLoader = loader.NewDwebDocumentLoader(api)
	expanded, err := proc.Expand(doc, opts)
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.MarshalIndent(expanded, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(b))
}
