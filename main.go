package main

import (
	"io"
	"net/http"
	"os"

	"github.com/jbenet/go-ipfs/core"
	"github.com/jbenet/go-ipfs/core/coreunix"
	"github.com/jbenet/go-ipfs/repo/fsrepo"
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

func main() {
	builder := core.NewNodeBuilder().Online()

	home := os.Getenv("HOME")
	r := fsrepo.At(home + "/.go-ipfs")
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
	http.ListenAndServe(":8080", nil)
	cancel()
}
