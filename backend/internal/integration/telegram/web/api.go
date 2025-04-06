package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celestix/gotgproto"
	"golang.org/x/sync/errgroup"

	"goout/config"
)

func Start(config *config.Config, errgroup *errgroup.Group, wa *webAuth) func(context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", wa.setInfo)
	mux.HandleFunc("/getAuthStatus", wa.getAuthStatus)

	server := &http.Server{
		Addr:    config.AuthAddr,
		Handler: mux,
	}

	errgroup.Go(server.ListenAndServe)

	return server.Shutdown
}

func (wa *webAuth) getAuthStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, wa.authStatus.Event)
}

// setInfo handle user info, set phone, code or passwd
func (wa *webAuth) setInfo(w http.ResponseWriter, req *http.Request) {
	action := req.URL.Query().Get("set")

	switch action {

	case "phone":
		fmt.Println("Rec phone")
		num := req.URL.Query().Get("phone")
		phone := "+" + num
		wa.ReceivePhone(phone)
		for wa.authStatus.Event == gotgproto.AuthStatusPhoneAsked ||
			wa.authStatus.Event == gotgproto.AuthStatusPhoneRetrial {
			continue
		}
	case "code":
		fmt.Println("Rec code")
		code := req.URL.Query().Get("code")
		wa.ReceiveCode(code)
		for wa.authStatus.Event == gotgproto.AuthStatusPhoneCodeAsked ||
			wa.authStatus.Event == gotgproto.AuthStatusPhoneCodeRetrial {
			continue
		}
	case "passwd":
		passwd := req.URL.Query().Get("passwd")
		wa.ReceivePasswd(passwd)
		for wa.authStatus.Event == gotgproto.AuthStatusPasswordAsked ||
			wa.authStatus.Event == gotgproto.AuthStatusPasswordRetrial {
			continue
		}
	}
	w.Write([]byte(wa.authStatus.Event))
}
