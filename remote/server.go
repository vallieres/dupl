package remote

import (
	"errors"
	"log"
	"net"
	"net/rpc"

	"fm.tul.cz/dupl/job"
	"fm.tul.cz/dupl/suffixtree"
	"fm.tul.cz/dupl/syntax"
)

type Dupl struct {
	stree    *suffixtree.STree
	schan    chan []*syntax.Node
	mchan    <-chan suffixtree.Match
	done     chan bool
	finished bool
}

func (d *Dupl) UpdateTree(seq []*syntax.Node, ignore *bool) error {
	if d.finished {
		return errors.New("suffix tree has been finished")
	}
	d.schan <- seq
	return nil
}

func (d *Dupl) NextMatch(threshold int, r *Response) error {
	if !d.finished {
		d.finished = true
		close(d.schan)
		<-d.done
		d.stree.Update(&syntax.Node{Type: -1})
		d.mchan = d.stree.FindDuplOver(threshold)
	}
	m, ok := <-d.mchan
	r.Match, r.Done = syntax.GetNodes(d.stree, m), !ok
	return nil
}

type Response struct {
	Match [][]*syntax.Node
	Done  bool
}

func RunServer(port string) {
	d := new(Dupl)
	rpc.Register(d)

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("error:", err)
	}
	log.Println("server started")

	for {
		if conn, err := l.Accept(); err != nil {
			log.Fatal(err.Error())
		} else {
			log.Println("connection accepted")
			d.finished = false
			d.schan = make(chan []*syntax.Node)
			d.stree, d.done = job.BuildTree(d.schan)

			rpc.ServeConn(conn)
			log.Println("done")
		}
	}
}
