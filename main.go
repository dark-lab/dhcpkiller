package main

import (
	"crypto/rand"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"

	"log"
	"net"
)

func main() {
	var err error

	m, err := net.ParseMAC("08-00-27-00-A8-E8")
	if err != nil {
		log.Printf("MAC Error:%v\n", err)
	}

	//Create a connection to use
	//We need to set the connection ports to 1068 and 1067 so we don't need root access
	c, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 68}), dhcp4client.SetRemoteAddr(net.UDPAddr{IP: net.IPv4bcast, Port: 67}))

	if err != nil {
		log.Println("Client Conection Generation:" + err.Error())
	}

	exampleClient, err := dhcp4client.New(dhcp4client.HardwareAddr(m), dhcp4client.Connection(c))
	if err != nil {
		log.Fatalf("Error:%v\n", err)
	}
	_, err = SendDHCPDeclinePacket(exampleClient, m, net.IPv4(192, 168, 1, 250))
	if err != nil {
		log.Fatalf("Couldn't send DeclinePacket:" + err.Error())
	} else {
		log.Println("Sent!!!")
	}

}

func DHCPDeclinePacket(hw net.HardwareAddr, ip net.IP) dhcp4.Packet {
	messageid := make([]byte, 4)
	if _, err := rand.Read(messageid); err != nil {
		panic(err)
	}

	packet := dhcp4.NewPacket(dhcp4.BootRequest)
	packet.SetCHAddr(hw)
	packet.SetCIAddr(ip)

	packet.SetXId(messageid)
	packet.SetBroadcast(true)

	packet.AddOption(dhcp4.OptionDHCPMessageType, []byte{byte(dhcp4.Decline)})
	return packet
}

func SendDHCPDeclinePacket(c *dhcp4client.Client, hw net.HardwareAddr, ip net.IP) (dhcp4.Packet, error) {
	DeclinePacket := DHCPDeclinePacket(hw, ip)
	DeclinePacket.PadToMinSize()

	return DeclinePacket, c.SendPacket(DeclinePacket)
}
