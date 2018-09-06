package main

import (
  "github.com/TheThingsNetwork/go-account-lib/account"
  "github.com/TheThingsNetwork/api/discovery"
  "github.com/TheThingsNetwork/api/discovery/discoveryclient"
  "github.com/TheThingsNetwork/api/gateway"
  "github.com/TheThingsNetwork/api/protocol"
  "github.com/TheThingsNetwork/api/protocol/lorawan"
  "github.com/TheThingsNetwork/api/router"
  "github.com/TheThingsNetwork/api/router/routerclient"
  "fmt"
  "time"
  "github.com/TheThingsNetwork/ttn/api/pool"
  "os"
)

const (
  accountServer = "https://account.thethingsnetwork.org"
  discoveryServer = "discovery.thethings.network:1900"
  routerID = "ttn-router-eu"
)

var (
  gatewayID string
  gatewayKey string
)

func createUplinkMessage(payload []byte) *router.UplinkMessage {
  // Creating a dummy uplink message, using the protocol buffer-generated types
  return &router.UplinkMessage{
    GatewayMetadata: gateway.RxMetadata{
      RSSI: -35,
      SNR:  5,
    },
    Payload: payload,
    ProtocolMetadata: protocol.RxMetadata{
      Protocol: &protocol.RxMetadata_LoRaWAN{
        LoRaWAN: &lorawan.Metadata{
          CodingRate: "4/5",
          DataRate:   "SF7BW125",
          Modulation: lorawan.Modulation_LORA,
        },
      },
    },
  }
}

func connectAndSendUplink(uplink *router.UplinkMessage) error {
  fmt.Printf("%+v\n", uplink)

  // Connecting to the TTN account server to fetch a token
  gwAccount := account.NewWithKey(accountServer, gatewayKey)
  gw, err := gwAccount.FindGateway(gatewayID)
  if err != nil {
    return err
  }
  fmt.Printf("%+v\n", gw)

  token, err := gwAccount.GetGatewayToken(gatewayID)
  if err != nil {
    return err
  }
  fmt.Printf("%+v\n", token)

  // Connecting to the TTN discovery server to get a connection to the router
  discoveryClient, err := discoveryclient.NewClient(discoveryServer, &discovery.Announcement{ID: gatewayID}, func() string { return "" })
  if err != nil {
    return err
  }

  // Connecting to the router
  routerAccess, err := discoveryClient.Get("router", routerID)
  if err != nil {
    return err
  }
  fmt.Printf("Router:\n%+v\n", routerAccess)

//  time.Sleep(5*time.Second)

//  apipool := pool.NewPool(context.Background(), pool.DefaultDialOptions.append())
  conn, err := pool.Global.DialSecure(routerAccess.NetAddress, nil)
  //conn, err := pool.Global.DialInsecure(routerAccess.NetAddress)
  //conn, err := pool.Global.DialInsecure(routerAccess.MqttAddress)
    if err != nil {
      return err
  }
  fmt.Printf("conn:\n%+v\n", conn)

  //routerConn := routerAccess.Dial()
  routerClient := routerclient.NewClient(routerclient.DefaultClientConfig)
  routerClient.AddServer("eu", conn)
  fmt.Printf("routerClient:\n%+v\n", routerClient)
  genericStream := routerClient.NewGatewayStreams(gatewayID, token.AccessToken, false)
  fmt.Printf("%+v\n", genericStream)

  genericStream.Uplink(uplink)
  // Sending the uplink
//  if err := genericStream.Uplink(uplink); err != nil {
//    return err
//  }
  // Uplink sent successfully!
  time.Sleep(time.Second)
  genericStream.Close()
  conn.Close()
  fmt.Printf("Program end\n")
  return nil
}

func main(){
  gatewayID = os.Getenv("GWNAME")
  gatewayKey = os.Getenv("GWKEY")

  message := createUplinkMessage([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
  err := connectAndSendUplink(message)
  if err != nil {
    fmt.Printf("Send failed with err: %s\n",err)
  }
}
