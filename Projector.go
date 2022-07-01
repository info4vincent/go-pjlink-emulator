package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	POWER_OFF     = 0
	POWER_ON      = 1
	POWER_COOLING = 2
	POWER_WARMING = 3

	AVMUTE_NONE  = 30
	AVMUTE_VIDEO = 11
	AVMUTE_AUDIO = 21
	AVMUTE_BOTH  = 31

	AVMUTE_UNMUTE_VIDEO = 10
	AVMUTE_UNMUTE_AUDIO = 20
	AVMUTE_UNMUTE_BOTH  = 30

	INPUT_RGB_1 = 11
	INPUT_RGB_2 = 12
	INPUT_RGB_3 = 13
	INPUT_RGB_4 = 14
	INPUT_RGB_5 = 15
	INPUT_RGB_6 = 16
	INPUT_RGB_7 = 17
	INPUT_RGB_8 = 18
	INPUT_RGB_9 = 19

	INPUT_VIDEO_1 = 21
	INPUT_VIDEO_2 = 22
	INPUT_VIDEO_3 = 23
	INPUT_VIDEO_4 = 24
	INPUT_VIDEO_5 = 25
	INPUT_VIDEO_6 = 26
	INPUT_VIDEO_7 = 27
	INPUT_VIDEO_8 = 28
	INPUT_VIDEO_9 = 29

	INPUT_DIGITAL_1 = 31 // HDMI 1
	INPUT_DIGITAL_2 = 32 // HDMI 2
	INPUT_DIGITAL_3 = 33 // DISPLAY PORT 1
	INPUT_DIGITAL_4 = 34 // DISPLAY PORT 2
	INPUT_DIGITAL_5 = 35 // HDBaseT
	INPUT_DIGITAL_6 = 36 // SDI
	INPUT_DIGITAL_7 = 37 // Not available for device
	INPUT_DIGITAL_8 = 38 // Not available for device
	INPUT_DIGITAL_9 = 39 // Not available for device

	INPUT_STORAGE_1 = 41
	INPUT_STORAGE_2 = 42
	INPUT_STORAGE_3 = 43
	INPUT_STORAGE_4 = 44
	INPUT_STORAGE_5 = 45
	INPUT_STORAGE_6 = 46
	INPUT_STORAGE_7 = 47
	INPUT_STORAGE_8 = 48
	INPUT_STORAGE_9 = 49

	INPUT_NETWORK_1 = 51
	INPUT_NETWORK_2 = 52
	INPUT_NETWORK_3 = 53
	INPUT_NETWORK_4 = 54
	INPUT_NETWORK_5 = 55
	INPUT_NETWORK_6 = 56
	INPUT_NETWORK_7 = 57
	INPUT_NETWORK_8 = 58
	INPUT_NETWORK_9 = 59
)

// Projector store
type Projector struct {
	_PJLinkUseAuthentication bool

	_PJLinkPower     int
	_PJLinkInput     int
	_PJLinkAVMute    int
	_PJLinkError     int
	_PJLinkLampHours int
	_PJLinkName      string
	_port            int

	_deviceCreatedAtTime time.Time
}

// NewProjector instance with defaults
func NewProjector() Projector {
	rand.Seed(time.Now().UnixNano())
	generatedName := "Emulator " + fmt.Sprint(rand.Intn(999-1)+1)

	projector := Projector{}
	projector._PJLinkName = generatedName
	projector._PJLinkPower = POWER_OFF
	projector._PJLinkInput = INPUT_DIGITAL_1
	projector._PJLinkAVMute = AVMUTE_UNMUTE_BOTH
	projector._PJLinkLampHours = 30000
	projector._port = 4352

	projector._deviceCreatedAtTime = time.Now()

	return projector
}

type cache struct {
	data map[string]string
	*sync.RWMutex
}

var c = cache{data: make(map[string]string), RWMutex: &sync.RWMutex{}}

// When a invalid PJLink command is received (Projector/Display failure)
// TODO (IMplement according PJLink spec)
var InvalidCommand = []byte("Invalid Command") // = ERR 4

func main() {

	log.SetOutput(os.Stdout)
	aProjector := NewProjector()
	log.Println("Started emulating a PJLink device (projector/display) with Name : " + aProjector._PJLinkName)
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(aProjector._port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept Error", err)
			continue
		}

		log.Println("Accepted ", conn.RemoteAddr())
		conn.Write([]byte("PJLINK 0\r"))

		//create a routine dont block
		go handleConnection(conn, aProjector)
	}

}

func handleConnection(conn net.Conn, projector Projector) {
	defer conn.Close()

	s := bufio.NewReader(conn)
	for {
		data, err := s.ReadString('\r')

		if err != nil {
			return
		}

		if data == "" {
			conn.Write([]byte(">"))
			//continue
			return
		}

		if data == "exit" {
			return
		}

		handleCommand(data, conn, &projector)
	}
}

func handleCommand(inp string, conn net.Conn, projector *Projector) {

	if len(inp) <= 0 || inp[0] != '%' {
		conn.Write(InvalidCommand)
		return
	}

	command := strings.TrimRight(inp, "\r") //remove '\r'

	switch command {
	case "%1POWR ?":
		get(command, fmt.Sprint(projector._PJLinkPower), conn)
	case "%1POWR 1":
		projector._PJLinkPower = POWER_ON
		replyOK(command, conn)
	case "%1POWR 0":
		projector._PJLinkPower = POWER_OFF
		replyOK(command, conn)
	case "%1NAME ?":
		get(command, fmt.Sprint(projector._PJLinkName), conn)
	case "%1LAMP ?":
		hoursInUse := math.Round(time.Now().Sub(projector._deviceCreatedAtTime).Seconds())
		remainingHours := projector._PJLinkLampHours - int(math.Mod(hoursInUse, float64(projector._PJLinkLampHours)))
		get(command, fmt.Sprint(remainingHours), conn)
	case "%1INPT ?":
		get(command, fmt.Sprint(projector._PJLinkInput), conn)
	default:
		if strings.HasPrefix(command, "%1INPT ") == true {
			newInputSource, _ := strconv.Atoi(strings.TrimPrefix(command, "%1INPT "))
			projector._PJLinkInput = newInputSource
			replyOK(command, conn)
			break
		}

		conn.Write(InvalidCommand)
	}

	conn.Write([]byte("\n>"))
}

func replyOK(cmd string, conn net.Conn) {
	str := strings.Split(cmd, " ")
	conn.Write([]byte(str[0] + "= OK"))
}

func get(cmd string, value string, conn net.Conn) {

	if len(cmd) < 1 {
		conn.Write(InvalidCommand)
		return
	}

	ret := strings.Replace(cmd, " ?", "="+value+"\r", 1)

	conn.Write([]byte(ret))
}
