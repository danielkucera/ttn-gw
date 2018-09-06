package main

import (
  "github.com/TheThingsNetwork/go-account-lib/account"
  "github.com/TheThingsNetwork/api/discovery"
  "github.com/TheThingsNetwork/api/gateway"
  "github.com/TheThingsNetwork/api/protocol"
  "github.com/TheThingsNetwork/api/protocol/lorawan"
  "github.com/TheThingsNetwork/api/router"
)

const (
  gatewayID = "rtl-sdr-test-gw"
  gatewayKey = "ttn-account-v2.-RsLrk3F0IWrYWBTACgLoJ6Cs-RMk3fAnSXkrElLenft8m-SH92UoxTAf341FsgavI5T_cE8sC--5DTmQFXXJw"

  accountServer = "https://account.thethingsnetwork.org"
  discoveryServer = "discovery.thethings.network:1900"
  routerID = "ttn-router-eu"
)

func createUplinkMessage(payload []byte) *router.UplinkMessage {
  // Creating a dummy uplink message, using the protocol buffer-generated types
  return &router.UplinkMessage{
    GatewayMetadata: &gateway.RxMetadata{
      Rssi: -35,
      Snr:  5,
    },
    Payload: payload,
    ProtocolMetadata: &protocol.RxMetadata{
      Protocol: &protocol.RxMetadata_Lorawan{
        Lorawan: &lorawan.Metadata{
          CodingRate: "4/5",
          DataRate:   "SF7BW125",
          Modulation: lorawan.Modulation_LORA,
        },
      },
    },
  }
}

func connectAndSendUplink(uplink *router.UplinkMessage) error {
  // Connecting to the TTN account server to fetch a token
  gwAccount := account.NewWithKey(accountServer, gatewayKey)
  gw, err := gwAccount.FindGateway(gatewayID)
  if err != nil {
    return err
  }
  token := gw.Token.AccessToken

  // Connecting to the TTN discovery server to get a connection to the router
  discoveryClient, err := discovery.NewClient(discoveryServer, &discovery.Announcement{Id: gatewayID}, func() string { return "" })
  if err != nil {
    return err
  }

  // Connecting to the router
  routerAccess, err := discoveryClient.Get("router", routerID)
  if err != nil {
    return err
  }

  routerConn := routerAccess.Dial()
  routerClient := router.NewRouterClientForGateway(router.NewRouterClient(c.routerConn), gatewayID, token)
  uplinkStream := router.NewMonitoredUplinkStream(routerClient)

  // Sending the uplink
  if err := uplinkStream.Send(uplink); err != nil {
    return err
  }
  // Uplink sent successfully!
  return nil
}


