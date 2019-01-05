package player

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
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
	PlayerArgument
	Location    Location    `json:"location"`
	Credentials Credentials `json:"credentials"`
}

type PlayerArgument struct {
	ID            string `json:"_id"`
	Rev           string `json:"_rev,omitempty"`
	Name          string `json:"name"`
	FavoriteColor string `json:"favoriteColor"`
}

func New(url, certpath, jwt string) PlayerService {
	return PlayerService{url, certpath, jwt}
}

func (p *PlayerService) GetAccounts() ([]Player, error) {
	var players []Player
	if response, status, err := p.doGet(strings.Join([]string{p.url, "accounts"}, "/")); err == nil {
		if status != 200 {
			err = errors.New("Bad Status")
		} else {
			err = json.Unmarshal(response, &players)
		}
		if err == nil {
			return players, nil
		} else {
			fmt.Println("Failed to parse response ", string(response))
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (p *PlayerService) Exists(id string) (bool, error) {
	if _, status, err := p.doGet(strings.Join([]string{p.url, "accounts", id}, "/")); err == nil {
		switch status {
		case 200:
			return true, nil
		case 404:
			return false, nil
		default:
			return false, fmt.Errorf("Unknown status %d", status)
		}
	} else {
		return false, err
	}
}

func (p *PlayerService) GetAccount(id string) (Player, error) {
	var player Player
	if response, status, err := p.doGet(strings.Join([]string{p.url, "accounts", id}, "/")); err == nil {
		if status != 200 {
			err = errors.New("Bad Status")
		} else {
			err = json.Unmarshal(response, &player)
		}
		if err == nil {
			return player, nil
		} else {
			return Player{}, err
		}
	} else {
		return Player{}, err
	}
}

func (p *PlayerService) CreatePlayer(arg PlayerArgument) (Player, error) {
	var player Player
	if body, err := json.Marshal(arg); err == nil {
		if response, status, err := p.doPost(strings.Join([]string{p.url, "accounts"}, "/"), body); err == nil {
			if err != nil {
				return Player{}, err
			}
			switch status {
			case 200, 201:
				{
					err = json.Unmarshal(response, &player)
					if err != nil {
						return Player{}, err
					} else {
						return player, nil
					}
				}
			default:
				return Player{}, fmt.Errorf("Bad Status %d", status)
			}
		} else {
			return Player{}, err
		}
	} else {
		return Player{}, err
	}
}

func (p *PlayerService) DeletePlayer(id string) (bool, error) {
	if _, status, err := p.doDelete(strings.Join([]string{p.url, "accounts", id}, "/")); err == nil {
		switch status {
		case 204, 200:
			return true, nil
		case 403:
			return false, errors.New("Unauthorized")
		case 404:
			return false, nil
		default:
			return false, fmt.Errorf("Unknown status %d", status)
		}
	} else {
		return false, err
	}

}

func (p *PlayerService) getClient() *http.Client {
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

func (p *PlayerService) doGet(url string) ([]byte, int, error) {
	return p.doInvoke("GET", url, nil)
}

func (p *PlayerService) doDelete(url string) ([]byte, int, error) {
	return p.doInvoke("DELETE", url, nil)
}

func (p *PlayerService) doPost(url string, body []byte) ([]byte, int, error) {
	return p.doInvoke("POST", url, body)
}

func (p *PlayerService) doInvoke(method, url string, body []byte) ([]byte, int, error) {
	var client = p.getClient()
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
	if p.jwt != "" {
		req.Header.Add("gameon-jwt", p.jwt)
	}
	req.Header.Set("Content-Type", "application/json")
	if response, err := client.Do(req); err == nil {
		if buf, err := ioutil.ReadAll(response.Body); err == nil {
			return buf, response.StatusCode, nil
		} else {
			return nil, response.StatusCode, err
		}
	} else {
		return nil, -1, err
	}
}
