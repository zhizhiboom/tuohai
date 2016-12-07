package im_api

import (
	"fmt"
	"net"
	"os"

	"tuohai/im_api/models"
	"tuohai/internal/svc"
	"tuohai/internal/util"
)

type ImApi struct {
	Opts         *Options
	httpListener net.Listener
	waitGroup    util.WaitGroupWrapper
	exitChan     chan int
}

func New(opts *Options) *ImApi {
	return &ImApi{Opts: opts, exitChan: make(chan int)}
}

func (api *ImApi) Main() {
	httpListener, err := net.Listen("tcp", api.Opts.HTTPAddress)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	models.InitDB(api.Opts.MysqlOptions())
	models.InitRedis(api.Opts.RedisHost, api.Opts.RedisPwd)

	api.httpListener = httpListener
	fmt.Println("LISTEN: ", api.httpListener.Addr().String())
	fmt.Println("PID: ", api.Opts.ID)

	api.waitGroup.Wrap(func() {
		svc.HttpService(api.httpListener, newHTTPServer())
	})

	api.waitGroup.Wrap(func() {
		api.friendLoop()
	})
}

func (api *ImApi) Close() {
	api.exitChan <- 0
	close(api.exitChan)
}
