package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jbenet/go-ipfs/core"
	"github.com/jbenet/go-ipfs/core/coreunix"
	"github.com/jbenet/go-ipfs/importer"
	"github.com/jbenet/go-ipfs/importer/chunk"
	"github.com/jbenet/go-ipfs/repo/fsrepo"
	uio "github.com/jbenet/go-ipfs/unixfs/io"
	u "github.com/jbenet/go-ipfs/util"

	"code.google.com/p/go.net/context"
)

var gnode *core.IpfsNode

func ServeIpfsRand(w http.ResponseWriter, r *http.Request) {
	read := io.LimitReader(u.NewTimeSeededRand(), 2048)

	str, err := coreunix.Add(gnode, read)
	if err != nil {
		w.WriteHeader(504)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(str))
	}
}

func ServeRandDir(w http.ResponseWriter, r *http.Request) {
	db := uio.NewDirectory(gnode.DAG)
	for i := 0; i < 50; i++ {
		read := io.LimitReader(u.NewTimeSeededRand(), 512)
		nd, err := importer.BuildDagFromReader(read, gnode.DAG, nil, chunk.DefaultSplitter)
		if err != nil {
			panic(err)
		}
		k, err := gnode.DAG.Add(nd)
		if err != nil {
			panic(err)
		}

		err = db.AddChild(fmt.Sprint(i), k)
		if err != nil {
			panic(err)
		}
	}

	nd := db.GetNode()
	k, err := gnode.DAG.Add(nd)
	if err != nil {
		w.WriteHeader(504)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(k.B58String()))
	}
}

func main() {
	builder := core.NewNodeBuilder().Online()

	r := fsrepo.At("~/.go-ipfs")
	if err := r.Open(); err != nil {
		panic(err)
	}

	builder.SetRepo(r)

	ctx, cancel := context.WithCancel(context.Background())
	node, err := builder.Build(ctx)
	if err != nil {
		panic(err)
	}

	gnode = node

	http.HandleFunc("/ipfsobject", ServeIpfsRand)
	http.HandleFunc("/ipfsdir", ServeRandDir)
	http.ListenAndServe(":8080", nil)
	cancel()
}
