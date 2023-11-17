package tofu

import (
	"context"
	"fmt"
	"net/http"
)

type Tofu struct {
	CTX context.Context

	*Graceful
	Config
	CorsConfig

	Models     *Models
	HTTPServer *http.Server
	DB         DBOperations

	MQTT *MQTT
}

func New(opts ...func(tofu *Tofu)) *Tofu {
	tf := &Tofu{}
	tf.CTX = context.Background()

	tf.Models = NewModels()

	tf.Graceful = NewGraceful(StopSignal())

	for _, opt := range opts {
		opt(tf)
	}

	return tf
}

func WithMySQLDB(config MySqlConfig) func(*Tofu) {
	return func(tofu *Tofu) {
		tofu.DB = NewMySqlDB(config, tofu.Models)
	}
}

func WithHTTPServer(httpConfig HTTPConfig, corsConfig CorsConfig) func(*Tofu) {
	return func(tofu *Tofu) {
		tofu.HTTPServer = &http.Server{}
		tofu.HTTPServer.Addr = fmt.Sprintf(httpConfig.Port)
		tofu.CorsConfig = corsConfig
	}
}

func WithMQTTBroker(config MQTTConfig) func(*Tofu) {
	return func(tofu *Tofu) {
		tofu.MQTT = NewMQTT(config)
	}
}

func (tofu *Tofu) Run() {
	if tofu.DB != nil {
		if err := tofu.DB.Migrate(); err != nil {
			panic(err)
		}
	}

	if tofu.HTTPServer != nil {
		api := NewHttpApi(tofu.Models, WithDatabase(tofu.DB))
		tofu.HTTPServer.Handler = api.GetHandler(tofu.CorsConfig)

		go func() {
			if err := tofu.HTTPServer.ListenAndServe(); err != nil {
				fmt.Println(err)
			}
		}()

		tofu.Graceful.GoNoErr(func() {
			if err := tofu.HTTPServer.Shutdown(tofu.CTX); err != nil && err != http.ErrServerClosed {
				fmt.Println("graceful shutdown http server error:", err)
				return
			}
			fmt.Println("grace down")
		})
		_ = tofu.Graceful.Wait()
	}

	if tofu.MQTT != nil && (len(tofu.MQTT.subscribers) > 0 || len(tofu.MQTT.publishers) > 0) {
		for _, subscriber := range tofu.MQTT.subscribers {
			go func(s SubFn) {
				tofu.MQTT.subscribe(s.Topic, s.fn)
			}(subscriber)
		}
		for _, publisher := range tofu.MQTT.publishers {
			go func(p PubFn) {
				tofu.MQTT.publish(p.Topic, p.fn)
			}(publisher)
		}
		tofu.Graceful.GoNoErr(func() {
			tofu.MQTT.disconnect()
		})
		_ = tofu.Graceful.Wait()
	}

}
