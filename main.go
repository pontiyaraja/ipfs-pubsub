package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	pubsub "github.com/pontiyaraja/ipfs-pubsub/pubsub"

	kiplog "github.com/KIPFoundation/kip-log"
)

func main() {
	sub, err := subscribe("COPY")

	done := make(chan bool)

	go func() {
		for {
			msg, err := sub.Next()
			if err != nil {
				panic(err)
			}
			// if copyData(msg.Data) != nil {
			// 	fmt.Println("Error copying data ", string(msg.Data))
			// }
			fmt.Println(string(msg.Data), "     ", msg.From.Pretty(), "        ", string(msg.Seqno), "        ", msg.TopicIDs)
		}
	}()
	for i := 0; i < 10; i++ {
		if err = publish("COPY", `{"hash":`+strconv.Itoa(i)+`}`); err != nil {
			panic(err)
		}
		// if err = publish("COPY", `{"hash":"QmWDLfZbA2fGuTiJLWjj88nG8dKiwG236KocUeS4LGFyFg","path":"/pandi/removedAKfile.pdf"}`); err != nil {
		// 	panic(err)
		// }
	}
	<-done
}

func publish(topic, data string) error {
	url := "http://localhost:5001/api/v0/pubsub/pub?arg=" + topic + "&arg=" + data
	resp, err := doRequestP(url, http.MethodGet)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
	return err
}

func subscribe(topic string) (*pubsub.PubSubSubscription, error) {
	url := "http://localhost:5001/api/v0/pubsub/sub?arg=" + topic + "&discover=true"
	readCloser, err := doRequest(url, http.MethodPost)
	if err != nil {
		panic(err)
	}
	return pubsub.NewPubSubSubscription(readCloser), err
}

func copyData(msg []byte) error {
	cpy := struct {
		Hash string `json:"hash"`
		Path string `json:"path"`
		Type string `json:"type"`
	}{}
	err := json.NewDecoder(bytes.NewReader(msg)).Decode(&cpy)
	if err != nil {
		return err
	}

	//http://localhost:5001/api/v0/files/cp?arg=<source>&arg=<dest>
	resp, err := doRequestP("http://localhost:5001/api/v0/files/cp?arg=/ipfs/"+cpy.Hash+"&arg="+cpy.Path, http.MethodPost)
	if err != nil {
		return err
	}
	fmt.Println("COPY Success Resp Data  : ", string(resp))
	return nil
}

func doRequest(URL, method string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		kiplog.KIPError("", err, nil)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		kiplog.KIPError("", err, nil)
		return nil, err
	}

	fmt.Println(resp.Status, resp.StatusCode)
	//defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return nil, errors.New("got wrong status from server")
	// }
	// respByte, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	kiplog.KIPError("", err, nil)
	// 	return nil, err
	// }
	// return respByte, err

	return resp.Body, err
}

func doRequestP(URL, method string) ([]byte, error) {
	fmt.Println(URL)
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		kiplog.KIPError("", err, nil)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		kiplog.KIPError("", err, nil)
		return nil, err
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status, resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("got wrong status from server")
	}
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		kiplog.KIPError("", err, nil)
		return nil, err
	}
	return respByte, err

	//return resp.Body, err
}
