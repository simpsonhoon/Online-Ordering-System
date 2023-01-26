package main

import (
	"context"
	"flag"

	"fmt"
	conf "lecture/oos/conf"
	ctl "lecture/oos/controller"
	"lecture/oos/logger"
	"lecture/oos/model"
	rt "lecture/oos/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func main() {
	//model 모듈 선언
	var configFlag = flag.String("config", "./conf/config.toml", "toml file to use for configuration")
	flag.Parse()
	cf := conf.GetConfig(*configFlag)

	if err := logger.InitLogger(cf); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	logger.Debug("ready server....")

	if mod, err := model.NewModel(); err != nil {
		fmt.Println(err)
	} else if controller, err := ctl.NewCTL(mod); err != nil { //controller 모듈 설정
		fmt.Println(err)
	} else if rt, err := rt.NewRouter(controller); err != nil { //router 모듈 설정
		fmt.Println(err)
	} else {
		mapi := &http.Server{
			Addr:           ":8080",
			Handler:        rt.Idx(),
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		g.Go(func() error {
			return mapi.ListenAndServe()
		})

		stopSig := make(chan os.Signal) //chan 선언
		// 해당 chan 핸들링 선언, SIGINT, SIGTERM에 대한 메세지 notify
		signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)
		<-stopSig //메세지 등록
		logger.Warn("Shutdown Server ...")
		// 해당 context 타임아웃 설정, 5초후 server stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mapi.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			logger.Info("timeout 5 seconds.")
		}
		logger.Info("Server stop")

		if err := g.Wait(); err != nil {
			logger.Error(err)
		}
	}
}
