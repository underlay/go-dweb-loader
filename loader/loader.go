package loader

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	cid "github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	cbor "github.com/ipfs/go-ipld-cbor"
	core "github.com/ipfs/interface-go-ipfs-core"
	path "github.com/ipfs/interface-go-ipfs-core/path"
	ld "github.com/piprate/json-gold/ld"
)

// Compile-time type check
var _ ld.DocumentLoader = (*DwebDocumentLoader)(nil)

// DwebDocumentLoader is an implementation of ld.DocumentLoader
// for ipfs:// and dweb:/ipfs/ URIs that an core.CoreAPI
type DwebDocumentLoader struct {
	api core.CoreAPI
}

// LoadDocument returns a RemoteDocument containing the contents of the
// JSON-LD resource from the given URL.
func (dl *DwebDocumentLoader) LoadDocument(uri string) (*ld.RemoteDocument, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
	}

	// Don't do anything with contextURL.
	var contextURL string

	var origin, path string
	if u.Scheme == "ipfs" {
		return dl.loadDocumentIPFS(uri, contextURL, u.Host, u.Path)
	} else if u.Scheme == "dweb" {
		if u.Path[:6] == "/ipfs/" {
			index := strings.Index(u.Path[6:], "/")
			if index == -1 {
				index = len(u.Path)
			} else {
				index += 6
			}
			origin = u.Path[6:index]
			path = u.Path[index:]
			return dl.loadDocumentIPFS(uri, contextURL, origin, path)
		} else if u.Path[:6] == "/ipld/" {
			return dl.loadDocumentIPLD(uri, contextURL, u.Path[6:])
		} else {
			err := "Unsupported dweb path: " + u.Path
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
	} else {
		err := "Unsupported URI scheme: " + u.Scheme
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
	}
}

func (dl *DwebDocumentLoader) loadDocumentIPLD(uri string, contextURL string, origin string) (*ld.RemoteDocument, error) {
	dagAPI := dl.api.Dag()
	var data []byte

	c, err := cid.Decode(origin)
	if err != nil {
		return nil, err
	}

	if c.Type() == cid.DagCBOR {
		node, err := dagAPI.Get(context.Background(), c)
		if err != nil {
			return nil, err
		}
		cborNode, isCborNode := node.(*cbor.Node)
		if !isCborNode {
			err := "Unsupported IPLD CID format: " + origin
			return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
		}
		data, err = cborNode.MarshalJSON()
		if err != nil {
			return nil, err
		}
	} else if c.Type() == cid.Raw {
		node, err := dagAPI.Get(context.Background(), c)
		if err != nil {
			return nil, err
		}
		data = node.RawData()
	} else {
		err := "Unsupported IPLD CID format: " + origin
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, err)
	}

	var document interface{}
	err = json.Unmarshal(data, &document)
	if err != nil {
		return nil, err
	}

	return &ld.RemoteDocument{DocumentURL: uri, Document: document, ContextURL: contextURL}, nil
}

func (dl *DwebDocumentLoader) loadDocumentIPFS(uri string, contextURL string, origin string, remainder string) (*ld.RemoteDocument, error) {
	c, err := cid.Decode(origin)
	if err != nil {
		return nil, err
	}

	unixfs := dl.api.Unixfs()
	root := path.IpfsPath(c)
	tail := path.Join(root, remainder)
	ctx := context.Background()
	node, err := unixfs.Get(ctx, tail)
	if err != nil {
		return nil, err
	} else if file, isFile := node.(files.File); isFile {
		document, err := ld.DocumentFromReader(file)
		if err != nil {
			return nil, err
		}
		return &ld.RemoteDocument{DocumentURL: uri, Document: document, ContextURL: contextURL}, nil
	} else {
		return nil, ld.NewJsonLdError(ld.LoadingDocumentFailed, "Cannot load directory")
	}
}

// NewDwebDocumentLoader creates a new instance of DwebDocumentLoader
func NewDwebDocumentLoader(api core.CoreAPI) *DwebDocumentLoader {
	return &DwebDocumentLoader{api}
}
