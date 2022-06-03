package httpserver

import (
	"fmt"
	"net"

	"github.com/muir/nject/nserve"
	"github.com/muir/nject/nvelope"
	"github.com/muir/nject/npoint"
	"github.com/muir/nfigure"
	"github.com/muir/nforce"
	"github.com/pkg/errors"
	"github.com/gorilla/mux"
)

type config struct {
	ListenAddress []string  `config:"ListenAddress" flags:"listen,split=space" default:":8000"`
}

type server struct {
	config
}

func New(app *nserve.App, creg *nfigure.Registry) *server {
	s := &server{}
	creg.Request(s.config)
	app.On(nserve.Start, func(app *nserve.App) error {
		if len(s.ListenAddress) == 0 {
			return errors.New("No listen address specified")
		}
		listeners := make([]net.Listener, 0, len(s.ListenAddress)) 
		var shutdownCalled bool
		shutdown := func() {
			shutdownCalled = true
			for _, listener := range listeners {
				_ = listenver.Close()
			}
		}
		var okay bool
		defer func() {
			if !okay { shutdown() }
		}
		handler := s.GetHandler()
		for _, addr := range s.ListenAddress {
			addr := addr
			listener, err := net.Listen("tcp", addr)
			if err != nil { return errors.Wrapf(err, "xm log recevier at %s", addr) }
			go func() {
				err := http.Serve(listener, handler)
				if err != nil && !shutdownCalled {
					panic(fmt.Sprintf("http.Serve for xm log recevier on %s exited early with %s", addr, err))
				}
			}
		}
		okay = true
		app.On(nserve.Stop, func() error {
			shutdown()
			return nil
		})
	})
	return s
}

func (s *server) GetHandler() http.Handler {
	r := mux.NewRouter()
	service := npoint.RegisterServiceWithMux("xm-receive", r,
		LOGGER,
		nvelope.InjectWriter,
		nvelope.EncodeJSON,
		nvelope.CatchPanic,
		nvelope.Nil204,
		nvelope.ReadBody,
		nvelope.DecodeJSON,
	)
	service.RegisterEndpoint("/injest", s.injest)
	return r
}
