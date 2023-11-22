package tofu

import (
	"context"
	"fmt"
	"net/http"

	"github.com/WojciechWiderski/tofu/tconfig"
	"github.com/WojciechWiderski/tofu/tdatabase"
	"github.com/WojciechWiderski/tofu/tdatabase/mysql"
	"github.com/WojciechWiderski/tofu/thelpers"
	"github.com/WojciechWiderski/tofu/thttp"
	"github.com/WojciechWiderski/tofu/tlogger"
	"github.com/WojciechWiderski/tofu/tmodel"
	"github.com/WojciechWiderski/tofu/tqueue"
)

type Tofu struct {
	CTX context.Context

	graceful   *thelpers.Graceful
	appConfig  tconfig.App
	corsConfig tconfig.Cors

	Models     *tmodel.Models
	HTTPServer *http.Server
	DB         tdatabase.DBOperations

	MQTT *tqueue.MQTT
}

func New(opts ...func(tofu *Tofu)) *Tofu {
	tf := &Tofu{}
	tf.CTX = context.Background()

	tf.Models = tmodel.NewModels()

	tf.graceful = thelpers.NewGraceful(thelpers.StopSignal())

	for _, opt := range opts {
		opt(tf)
	}

	return tf
}

func WithMySQLDB(config tconfig.MySql) func(*Tofu) {
	return func(tofu *Tofu) {
		tofu.DB = mysql.New(config, tofu.Models)
	}
}

func WithHTTPServer(httpConfig tconfig.HTTP, corsConfig tconfig.Cors) func(*Tofu) {
	tlogger.Info(fmt.Sprintf("Create http server"))
	return func(tofu *Tofu) {
		tofu.HTTPServer = &http.Server{}
		tofu.HTTPServer.Addr = fmt.Sprintf(httpConfig.Port)
		tofu.corsConfig = corsConfig
	}
}

func WithMQTTBroker(config tconfig.MQTT) func(*Tofu) {
	tlogger.Info(fmt.Sprintf("Create mqtt broker"))
	return func(tofu *Tofu) {
		tofu.MQTT = tqueue.NewMqtt(config)
	}
}

func (tofu *Tofu) Run() {
	if tofu.DB != nil {
		if err := tofu.DB.Migrate(); err != nil {
			tlogger.Error(fmt.Sprintf("tofu.DB.Migrate error! Error: %v", err))
			panic(err)
		}
	}

	if tofu.HTTPServer != nil {
		api := thttp.NewHttpApi(tofu.Models, thttp.WithDatabase(tofu.DB))
		tofu.HTTPServer.Handler = api.GetHandler(tofu.corsConfig)

		go func() {
			tlogger.Info(fmt.Sprintf("Http api listen on port: %s", tofu.HTTPServer.Addr))
			if err := tofu.HTTPServer.ListenAndServe(); err != nil {
				tlogger.Error(fmt.Sprintf(" tofu.HTTPServer.ListenAndServe error! Error: %v", err))
			}
		}()

		tofu.graceful.GoNoErr(func() {
			if err := tofu.HTTPServer.Shutdown(tofu.CTX); err != nil && err != http.ErrServerClosed {
				tlogger.Error(fmt.Sprintf("Graceful shutdown thttp server terror: %v", err))
				return
			}
			tlogger.Info("HttpApi grace down!")
		})
		_ = tofu.graceful.Wait()
	}

	if tofu.MQTT != nil && (len(tofu.MQTT.Subscribers) > 0 || len(tofu.MQTT.Publishers) > 0) {
		for _, subscriber := range tofu.MQTT.Subscribers {
			go func(s tqueue.SubFn) {
				tofu.MQTT.Subscribe(s.Topic, s.Fn)
			}(subscriber)
		}
		for _, publisher := range tofu.MQTT.Publishers {
			go func(p tqueue.PubFn) {
				tofu.MQTT.Publish(p.Topic, p.Fn)
			}(publisher)
		}
		tofu.graceful.GoNoErr(func() {
			tofu.MQTT.Disconnect()
		})
		_ = tofu.graceful.Wait()
	}

}
