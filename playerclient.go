package player

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type PlayerService struct {
	url      string
	certpath string
	jwt      string
}

type Location struct {
	Location string `json:"location"`
}

type Credentials struct {
	SharedSecret string `json:"sharedSecret"`
}

type Player struct {
	ID            string      `json:"_id"`
	Rev           string      `json:"_rev"`
	Name          string      `json:"name"`
	FavoriteColor string      `json:"favoriteColor"`
	Location      Location    `json:"location"`
	Credentials   Credentials `json:"credentials"`
}

func New(url, certpath, jwt string) PlayerService {
	return PlayerService{url, certpath, jwt}
}

func (p *PlayerService) GetAccounts() ([]Player, error) {
	var players []Player
	if response, err := doGet(p, strings.Join([]string{p.url, "accounts"}, "/")); err == nil {
		err = json.Unmarshal(response, &players)
		if err == nil {
			return players, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (p *PlayerService) GetAccount(id string) (Player, error) {
	var player Player
	if response, err := doGet(p, strings.Join([]string{p.url, "accounts", id}, "/")); err == nil {
		err = json.Unmarshal(response, &player)
		if err == nil {
			return player, nil
		} else {
			return Player{}, err
		}
	} else {
		return Player{}, err
	}
}

func getClient(p *PlayerService) *http.Client {
	var tr *http.Transport
	if p.certpath != "" {
		CAPool := x509.NewCertPool()
		severCert, err := ioutil.ReadFile(p.certpath)
		if err != nil {
			log.Fatal("Could not load server certificate!")
		}
		CAPool.AppendCertsFromPEM(severCert)
		config := &tls.Config{RootCAs: CAPool}
		tr = &http.Transport{TLSClientConfig: config}
	}
	var client *http.Client
	switch {
	case tr != nil:
		client = &http.Client{
			Timeout:   time.Second * 15,
			Transport: tr}
	default:
		client = &http.Client{
			Timeout: time.Second * 15,
		}
	}
	return client
}

func doGet(p *PlayerService, url string) ([]byte, error) {
	var client = getClient(p)
	req, _ := http.NewRequest("GET", url, nil)
	if p.jwt != "" {
		req.Header.Add("gameon-jwt", p.jwt)
	}
	if response, err := client.Do(req); err == nil {
		if buf, err := ioutil.ReadAll(response.Body); err == nil {
			return buf, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
