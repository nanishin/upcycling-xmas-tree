package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "log"
    "flag"
    "time"
    "math/rand"
    "encoding/json"
    "github.com/kellydunn/go-opc"
	"os"
	"os/exec"
	"strings"
	"strconv"
)   

type Color struct {
    R, G, B uint8
}

type Scroller struct {
    delay, train_len int
    random bool
    color Color
}

var color = Color{255,0,0}
var home_c chan Scroller

var proximity_sensor_control = 0
var led_color = "led_r" // RED led by default # led_r: RED, led_g: GREEN, led_b: BLUE
var led_file_path = "/sys/class/leds/led_r/brightness" // RED led file path by default

var fadecandy_control = 0

func random(min, max int) uint8 {
    xr := rand.Intn(max - min) + min
    return uint8(xr)
}

var serverPtr = flag.String("fcserver", "localhost:7890", "Fadecandy server and port to connect to")
var listenPortPtr = flag.Int("port", 8080, "Port to serve UI from")
var leds_len = flag.Int("leds", 128, "Number of LEDs in the string")

func main() {
    rand.Seed(time.Now().Unix())
    
    flag.Parse()
    
    home_c = make(chan Scroller, 1)

	// LCD OFF
	value := []byte("overwatch")
	ioutil.WriteFile("/sys/power/wake_lock", value, 0644)
	value = []byte("mem")
	ioutil.WriteFile("/sys/power/state", value, 0644)

    // Start proximity control
    proximity_sensor_control = 1
    value = []byte("1")
    ioutil.WriteFile("/sys/class/input/input4/enable", value, 0644) // On S3
	go StartProximityControl(200 * time.Millisecond) // 200ms
    
    fs := http.FileServer(http.Dir("/root/static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.Handle("/", http.StripPrefix("/", fs))
    http.HandleFunc("/update", UpdateHandler)

    log.Println("Listening on",fmt.Sprintf("http://0.0.0.0:%d",*listenPortPtr), "...")
    http.ListenAndServe(fmt.Sprintf(":%d",*listenPortPtr), nil)
}   

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

    // do these stupid hacks for parsing JSON. 
    // go is pretty bad at this
    body, _ := ioutil.ReadAll(r.Body)    
    var f interface{}
    var inscroll Scroller
    json.Unmarshal(body, &f)

    m := f.(map[string]interface{})

    inscroll.delay = int(m["delay"].(float64))
    inscroll.train_len = int(m["train_len"].(float64))
    inscroll.random = bool(m["random"].(bool))
    colormap := m["color"].(map[string]interface{})
    inscroll.color.R = uint8(colormap["r"].(float64))
    inscroll.color.G = uint8(colormap["g"].(float64))
    inscroll.color.B = uint8(colormap["b"].(float64))

    ss := inscroll
    
    //send on the home channel, nonblocking
    select {
        case home_c <- ss:
        default:
            log.Println("msg NOT sent")
    }

    fmt.Fprintf(w, "HomeHandler", ss.delay)
} 

func LEDSender( c chan Scroller, server string, leds_len int ) {
        
    props := Scroller{18, 18, true, Color{255,0,0}}
    props.delay = 100

     // Create a client
    oc := opc.NewClient()
    err := oc.Connect("tcp", server)
    if err != nil {
        log.Fatal("Could not connect to Fadecandy server", err)
    }

    for fadecandy_control == 1 { 
        for i := 0; i < leds_len; i++ {
            // send pixel data
            m := opc.NewMessage(0)
            m.SetLength(uint16(leds_len*3))
            
            for ii := 0; ii < props.train_len; ii++ {
                pix := i+ii
                if pix >= leds_len {
                    pix = props.train_len - ii - 1
                }
                if props.random {
                    m.SetPixelColor(pix, random(2,255), random(2,255), random(2,255))
                } else {
                    m.SetPixelColor(pix, props.color.R, props.color.G, props.color.B)
                }
            }
            
            err := oc.Send(m) 
            if err != nil {
                log.Println("couldn't send color",err)
            }             
            time.Sleep(time.Duration(props.delay) * time.Millisecond)

            // receive from channel
            select {
                case props = <- c:
                default:
            }
        }
    }

    for i := 0; i < leds_len; i++ {
        // send pixel data
        m := opc.NewMessage(0)
        m.SetLength(uint16(leds_len*3))
        
        for ii := 0; ii < props.train_len; ii++ {
            pix := i+ii
            if pix >= leds_len {
                pix = props.train_len - ii - 1
            }
            m.SetPixelColor(pix, 0, 0, 0)
        }
        
        err := oc.Send(m) 
        if err != nil {
            log.Println("couldn't send color",err)
        }             
    }
}

func StartProximityControl(delay time.Duration) {
	var pre_action  = "stop"
	for proximity_sensor_control == 1 {
		data, err := ioutil.ReadFile("/sys/class/sensors/proximity_sensor/state")
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadFile err : %v\n", err)
			continue
		}
		value := strings.Split(string(data), "\n")
		proximity_value, err := strconv.Atoi(value[0])
		if err != nil {
			fmt.Println(err)
			proximity_value = 0
		}

		if proximity_value >= 15  && pre_action == "stop" {
			pre_action = "start"
            value := []byte("100")
            led_file_path = fmt.Sprintf("/sys/class/leds/%s/brightness", led_color)
            ioutil.WriteFile(led_file_path, value, 0644)

            fadecandy_control = 1
            go func() { LEDSender(home_c, *serverPtr, *leds_len) }()

            PlaySound("rudolf_nose")
		} else if proximity_value < 15 && pre_action == "start" {
			pre_action = "stop"
            value := []byte("0")
            led_file_path = fmt.Sprintf("/sys/class/leds/%s/brightness", led_color)
            ioutil.WriteFile(led_file_path, value, 0644)

            fadecandy_control = 0
		} else {
			//fmt.Printf("Same Action!!!! Just Ignore!!!\n")
		}
		time.Sleep(delay)
	}
}

var cmd *exec.Cmd
func PlaySound(src string) {
	file_path := fmt.Sprintf("/root/%s.mp3", src)
	cmd = exec.Command("mpg321", file_path)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
}
