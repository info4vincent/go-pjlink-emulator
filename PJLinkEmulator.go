package main

import (
	"bufio"
	"flag"
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

// PJLinkDevice store
type PJLinkDevice struct {
	_PJLinkUseAuthentication bool

	_PJLinkPower     int
	_PJLinkInput     int
	_PJLinkAVMute    int
	_PJLinkError     int
	_PJLinkLampHours int
	_PJLinkName      string
	_PJLinkClass     int
	_port            int

	_deviceCreatedAtTime time.Time
	_coolingDownDuration time.Duration
	_warmingUpDuration   time.Duration
	_deviceThermalAtTime time.Time
	sync.Mutex           // wraps a synchronization flag
}

func (re *PJLinkDevice) turn_power_on() {
	re.Lock()
	defer re.Unlock()
	if re._warmingUpDuration == 0 {
		re._PJLinkPower = POWER_ON
	} else {
		re._PJLinkPower = POWER_WARMING
		re._deviceThermalAtTime = time.Now()
	}
}

func (re *PJLinkDevice) turn_power_off() {
	re.Lock()
	defer re.Unlock()
	if re._warmingUpDuration == 0 {
		re._PJLinkPower = POWER_OFF
	} else {
		re._PJLinkPower = POWER_COOLING
		re._deviceThermalAtTime = time.Now()
	}
}

func (re *PJLinkDevice) set_power_thermal_status() {
	re.Lock()
	defer re.Unlock()
	switch re._PJLinkPower {
	case POWER_ON:
	case POWER_OFF:
	case POWER_WARMING:
		if re._deviceThermalAtTime.Add(re._warmingUpDuration).Before(time.Now()) {
			re._PJLinkPower = POWER_ON
		}
	case POWER_COOLING:
		if re._deviceThermalAtTime.Add(re._coolingDownDuration).Before(time.Now()) {
			re._PJLinkPower = POWER_OFF
		}
	}
	log.Println("POWER state:", re._PJLinkPower)
	return
}

func (re *PJLinkDevice) set_input_source(newSource int) {
	re.Lock()
	defer re.Unlock()
	if newSource < INPUT_RGB_1 || newSource > INPUT_NETWORK_9 {
		log.Println("Error source is invalid:", newSource)
		return
	}
	re._PJLinkInput = newSource
	log.Println("SOURCE:", re._PJLinkInput)
	return

}

// NewProjector instance with defaults
func NewProjector() PJLinkDevice {
	rand.Seed(time.Now().UnixNano())
	generatedName := "Projector Emulator " + fmt.Sprint(rand.Intn(999-1)+1)

	projector := PJLinkDevice{}
	projector._PJLinkName = generatedName
	projector._PJLinkPower = POWER_OFF
	projector._PJLinkInput = INPUT_DIGITAL_1
	projector._PJLinkAVMute = AVMUTE_UNMUTE_BOTH
	projector._PJLinkLampHours = 30000
	projector._PJLinkClass = 2 // Projectors can also be of class 1
	projector._port = 4352

	projector._deviceCreatedAtTime = time.Now()
	projector._coolingDownDuration = time.Duration(12 * time.Second)
	projector._warmingUpDuration = time.Duration(6 * time.Second)

	return projector
}

func NewDisplay() PJLinkDevice {
	rand.Seed(time.Now().UnixNano())
	generatedName := "Display Emulator " + fmt.Sprint(rand.Intn(999-1)+1)

	display := PJLinkDevice{}
	display._PJLinkName = generatedName
	display._PJLinkPower = POWER_OFF
	display._PJLinkInput = INPUT_DIGITAL_1
	display._PJLinkAVMute = AVMUTE_UNMUTE_BOTH
	display._PJLinkLampHours = -1
	display._PJLinkClass = 1 // Can displays also be of class 2 ?
	display._port = 4352

	display._deviceCreatedAtTime = time.Now()
	display._coolingDownDuration = 0
	display._warmingUpDuration = 0

	return display
}

// When a invalid PJLink command is received (Projector/Display failure)
// TODO (IMplement according PJLink spec)
var InvalidCommand = []byte("Invalid Command") // = ERR 4

func main() {
	log.SetOutput(os.Stdout)

	isDisplayPtr := flag.Bool("display", false,
		"Emulate a display")
	flag.Parse()

	aDevice := NewProjector()
	if *isDisplayPtr == true {
		fmt.Print("Will emulate a display...")
		aDevice = NewDisplay()
	} else {
		fmt.Print("Will emulate a projector...")
	}

	log.Println("Started emulating a PJLink device (projector/display) with Name : " + aDevice._PJLinkName)
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(aDevice._port))
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

		go handleConnection(conn, &aDevice)
	}
}

func handleConnection(conn net.Conn, projector *PJLinkDevice) {
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

		handleCommand(data, conn, projector)
	}
}

func handleCommand(inp string, conn net.Conn, pjLinkDevice *PJLinkDevice) {

	if len(inp) <= 0 || inp[0] != '%' {
		conn.Write(InvalidCommand)
		return
	}

	command := strings.TrimRight(inp, "\r") //remove '\r'

	switch command {
	case "%1POWR ?":
		pjLinkDevice.set_power_thermal_status()
		get(command, fmt.Sprint(pjLinkDevice._PJLinkPower), conn)
	case "%1POWR 1":
		pjLinkDevice.turn_power_on()
		replyOK(command, conn)
	case "%1POWR 0":
		pjLinkDevice.turn_power_off()
		replyOK(command, conn)
	case "%1NAME ?":
		get(command, fmt.Sprint(pjLinkDevice._PJLinkName), conn)
	case "%1LAMP ?":
		if pjLinkDevice._PJLinkLampHours == -1 {
			// No lamp available:
			// Returning PJLink 'NoLamp' result (See README.md PJLink Class 2, PDF)
			replyERR1(command, conn)
		} else {
			hoursInUse := math.Round(time.Now().Sub(pjLinkDevice._deviceCreatedAtTime).Seconds())
			remainingHours := pjLinkDevice._PJLinkLampHours - int(math.Mod(hoursInUse, float64(pjLinkDevice._PJLinkLampHours)))
			get(command, fmt.Sprint(remainingHours), conn)
		}
	case "%1INPT ?":
		get(command, fmt.Sprint(pjLinkDevice._PJLinkInput), conn)
	case "%1CLSS ?":
		get(command, fmt.Sprint(pjLinkDevice._PJLinkClass), conn)
	default:
		if strings.HasPrefix(command, "%1INPT ") == true {
			newInputSource, _ := strconv.Atoi(strings.TrimPrefix(command, "%1INPT "))

			pjLinkDevice.set_input_source(newInputSource)
			replyOK(command, conn)
			break
		}

		conn.Write(InvalidCommand)
	}

	conn.Write([]byte("\n>"))
}

func replyOK(cmd string, conn net.Conn) {
	str := strings.Split(cmd, " ")
	conn.Write([]byte(str[0] + "=OK"))
}

func replyERR1(cmd string, conn net.Conn) {
	str := strings.Split(cmd, " ")
	conn.Write([]byte(str[0] + "=ERR1"))
}

func get(cmd string, value string, conn net.Conn) {

	if len(cmd) < 1 {
		conn.Write(InvalidCommand)
		return
	}

	ret := strings.Replace(cmd, " ?", "="+value+"\r", 1)

	conn.Write([]byte(ret))
}
