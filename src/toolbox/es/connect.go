package es

import (
	"fmt"
	"math"
	"net/http"
	"orla-alert/solte.lab/src/config"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

func InitConnect(conf *config.ElasticSearch) error {
	if conf.Main != nil {
		nrdConn, err := elasticsearch.NewClient(elasticsearch.Config{
			Username:  conf.Main.Auth.Username,
			Password:  conf.Main.Auth.Password,
			Addresses: conf.Main.URLS,
			Transport: &http.Transport{},
			RetryBackoff: func(i int) time.Duration {
				d := time.Duration(math.Exp(float64(i))) * time.Second
				logger := zap.L().Named("Elastic Search")
				logger.Error(fmt.Sprintf("connection lost retry %d in %s...", i, d))
				return d
			},
		})
		if err != nil {
			return err
		}
		conf.Main.Client = nrdConn
	}

	return nil
}
