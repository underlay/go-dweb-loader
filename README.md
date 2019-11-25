# go-ld-loader

Sometimes you want to be sure that the remote JSON-LD contexts you're loading are really the ones the author intended, and a good way of doing this is content-addressed contexts using `ipfs://` or `dweb:/ipfs/` URI schemes.

This module has two interfaces that satisfy the `ld.DocumentLoader` interface from `github.com/piprate/json-gold` - a `HTTPDocumentLoader` that uses the HTTP API interface from `github.com/ipfs/go-ipfs-api`, and a `CoreDocumentLoader` that uses the CoreAPI instance from `github.com/ipfs/interface-go-ipfs-core`, which is what you get passed if you write a [`go-ipfs` plugin](https://github.com/ipfs/go-ipfs/blob/master/docs/plugins.md).

Usage should be fairly straightforward.
