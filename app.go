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

func (t *Tofu) SetOwnDB(db tdatabase.DBOperations) {

	t.DB = db
}

func (t *Tofu) Run() {
	if t.DB != nil {
		if err := t.DB.Migrate(); err != nil {
			tlogger.Error(fmt.Sprintf("tofu.DB.Migrate error! Error: %v", err))
			panic(err)
		}
	}

	if t.HTTPServer != nil {
		api := thttp.NewHttpApi(t.Models, thttp.WithDatabase(t.DB))
		t.HTTPServer.Handler = api.GetHandler(t.corsConfig)

		go func() {
			tlogger.Info(fmt.Sprintf("Http api listen on port: %s", t.HTTPServer.Addr))
			if err := t.HTTPServer.ListenAndServe(); err != nil {
				tlogger.Error(fmt.Sprintf(" tofu.HTTPServer.ListenAndServe error! Error: %v", err))
			}
		}()

		t.graceful.GoNoErr(func() {
			if err := t.HTTPServer.Shutdown(t.CTX); err != nil && err != http.ErrServerClosed {
				tlogger.Error(fmt.Sprintf("Graceful shutdown thttp server terror: %v", err))
				return
			}
			tlogger.Info("HttpApi grace down!")
		})
		_ = t.graceful.Wait()
	}

	if t.MQTT != nil && (len(t.MQTT.Subscribers) > 0 || len(t.MQTT.Publishers) > 0) {
		for _, subscriber := range t.MQTT.Subscribers {
			go func(s tqueue.SubFn) {
				t.MQTT.Subscribe(s.Topic, s.Fn)
			}(subscriber)
		}
		for _, publisher := range t.MQTT.Publishers {
			go func(p tqueue.PubFn) {
				t.MQTT.Publish(p.Topic, p.Fn)
			}(publisher)
		}
		t.graceful.GoNoErr(func() {
			t.MQTT.Disconnect()
		})
		_ = t.graceful.Wait()
	}

}
