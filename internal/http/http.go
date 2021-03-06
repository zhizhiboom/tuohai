package http

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"tuohai/internal/console"
	"tuohai/internal/pb/IM_Message"
	"tuohai/internal/pb/MsgSrv"
)

func SendLogicMsg(ConnLogicRPCAddress string, p *IM_Message.IMMsgData) (*IM_Message.IMMsgDataAck, error) {
	conn, err := grpc.Dial(ConnLogicRPCAddress, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(10*time.Millisecond))
	if err != nil {
		console.StdLog.Error(err)
		log.Println(err)
		return nil, err
	}
	defer conn.Close()
	c := MsgSrv.NewMsgLogicClient(conn)

	r, err := c.SendMsg(context.Background(), p)
	if err != nil {
		log.Println(err)
		console.StdLog.Error(err)
		return r, err
	}
	return r, nil
}

func post(url string, payload []byte) ([]byte, error) {
	return client("POST", url, payload)
}

func client(method, url string, payload []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return body, nil
}

func get(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return body, nil
}

func Groups(URL string) ([]byte, error) {
	data, err := get(URL)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Users(URL string) ([]byte, error) {
	data, err := get(URL)
	if err != nil {
		return nil, err
	}
	return data, nil
}
