package main // Look README.md

import (
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"wm24_api/sql"

	"github.com/kataras/iris/v12"
)

func main() {

	// init app and load static resource
	app := iris.Default()
	tmpl := iris.HTML("./", ".html")
	app.RegisterView(tmpl)
	app.HandleDir("./", http.Dir("./"))
	// db, secret := init_db()
	// subRouter := api.Router(db, secret)
	// app.PartyFunc("/", subRouter)

	/*************************Custom Routers****************************/

	// for data api files
	api := app.Party("/api")
	{
		api.Use(iris.Compression)
		api.Get("/{action}", run_action)
		api.Post("/{action}", run_action)
	}

	// for static html files
	html := app.Party("/")
	{
		html.Use(iris.Compression)
		html.Get("/{page}", func(ctx iris.Context) {
			page := ctx.Params().Get("page")
			ctx.View(page)
		})
		// html.Post("/{action}", run_action)
	}

	/*************************Starting Server****************************/
	addr := fmt.Sprintf(":%s", getenv("PORT", "9001"))
	app.Listen(addr)
}

func init_db() (*sql.MySQL, string) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		getenv("MYSQL_USER", "root"),
		getenv("MYSQL_PASSWORD", "19897216"),
		getenv("MYSQL_HOST", "ccmeta.com"),
		getenv("MYSQL_DATABASE", "go_acs"),
	)

	db, err := sql.ConnectMySQL(dsn)
	if err != nil {
		log.Fatalf("error connecting to the MySQL database: %v", err)
	}

	secret := getenv("JWT_SECRET", "EbnJO3bwmX")
	return db, secret
}

func getenv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}

/*
<?xml version='1.0' encoding='utf-8' standalone='yes' ?>
<WifiConfigStoreData>
<int name="Version" value="3" />
<SoftAp>
<string name="WifiSsid">&quot;Wm24&quot;</string>
<boolean name="HiddenSSID" value="false" />
<int name="SecurityType" value="2" />
<string name="Passphrase">88888888</string>
<int name="MaxNumberOfClients" value="0" />
<boolean name="ClientControlByUser" value="false" />
<boolean name="AutoShutdownEnabled" value="true" />
<long name="ShutdownTimeoutMillis" value="-1" />
<BlockedClientList />
<AllowedClientList />
<boolean name="BridgedModeOpportunisticShutdownEnabled" value="true" />
<int name="MacRandomizationSetting" value="2" />
<BandChannelMap>
<BandChannel>
<int name="Band" value="1" />
<int name="Channel" value="0" />
</BandChannel>
</BandChannelMap>
<boolean name="80211axEnabled" value="true" />
<boolean name="UserConfiguration" value="true" />
<long name="BridgedModeOpportunisticShutdownTimeoutMillis" value="-1" />
<VendorElements />
<boolean name="80211beEnabled" value="true" />
<string name="PersistentRandomizedMacAddress">8a:5b:e0:9f:35:e5</string>
</SoftAp>
</WifiConfigStoreData>
*/
type WiFiXML struct {
	WifiConfigStoreData xml.Name `xml:"WifiConfigStoreData"`
	SoftAp              struct {
		XMLName    xml.Name `xml:"SoftAp"`
		AP_Strings []struct {
			NAME  string `xml:"name,attr"`
			VALUE string `xml:",chardata"`
		} `xml:"string"`
		AP_Ints []struct {
			NAME  string `xml:"name,attr"`
			VALUE string `xml:"value,attr"`
		} `xml:"int"`
		AP_BOOLS []struct {
			NAME  string `xml:"name,attr"`
			VALUE string `xml:"value,attr"`
		} `xml:"boolean"`
		AP_BandChannelMap struct {
			XMLName     xml.Name `xml:"BandChannelMap"`
			BandChannel struct {
				XMLName          xml.Name `xml:"BandChannel"`
				BandChannel_Ints []struct {
					NAME  string `xml:"name,attr"`
					VALUE string `xml:"value,attr"`
				} `xml:"int"`
			} `xml:"BandChannel"`
		} `xml:"BandChannelMap"`
	} `xml:"SoftAp"`
}

func run_action(ctx iris.Context) {

	action := ctx.Params().Get("action")
	switch action {

	case `get_device_info`:
		firmwarewVersion, _ := exec.Command("/system/bin/getprop", "ro.mediatek.version.release").Output()
		serialNumber, _ := exec.Command("/system/bin/getprop", "ro.serialno").Output()
		imei, _ := exec.Command("service", "call", "iphonesubinfo", "1").Output()

		wifi_text, err := exec.Command("cat", "/data/misc/apexdata/com.android.wifi/WifiConfigStoreSoftAp.xml").Output()
		if err != nil {
			ctx.StopWithError(500, err)
			return
		}
		wifi_obj := new(WiFiXML)
		if err = xml.Unmarshal(wifi_text, wifi_obj); err != nil {
			ctx.StopWithError(500, err)
			return
		}
		mac_addr := wifi_obj.SoftAp.AP_Strings[2].VALUE

		ctx.JSON(iris.Map{
			"result":           "ok",
			"serialNumber":     string(serialNumber),
			"imei":             string(imei),
			"imsi":             "00000000000000000",
			"hardwareVersion":  "1.0.0",
			"softwarewVersion": "随便自定义1_1_1 和WEBUI重复了？",
			"firmwarewVersion": string(firmwarewVersion),
			"webUIVersion":     "随便自定义1_1_1",
			"mac":              mac_addr,
			"wanIP":            "10.40.86.109",
		})
	case `get_pin_setting`:
		pinRemain, err1 := exec.Command("/system/bin/getprop", "vendor.gsm.sim.retry.pin1").Output()
		pinEnabled, err2 := exec.Command("/system/bin/getprop", "gsm.slot1.num.pin1").Output()
		pinStatus, err3 := exec.Command("/system/bin/getprop", "gsm.slot1.num.pin1").Output()
		_pinRemain, _ := strconv.Atoi(string(pinRemain[:len(pinRemain)-1]))
		_pinEnabled, _ := strconv.Atoi(string(pinEnabled[:len(pinEnabled)-1]))
		_pinStatus, _ := strconv.Atoi(string(pinStatus[:len(pinStatus)-1]))
		if err3 != nil || err2 != nil || err1 != nil {
			ctx.StopWithError(500, err3)
			ctx.StopWithError(500, err2)
			ctx.StopWithError(500, err1)
			return
		}
		ctx.JSON(iris.Map{
			"result":     "ok",
			"pinRemain":  _pinRemain,
			"pinEnabled": _pinEnabled,
			"pinStatus":  _pinStatus,
		})
	case `get_wifi_settings`:
		wifi_text, err := exec.Command("cat", "/data/misc/apexdata/com.android.wifi/WifiConfigStoreSoftAp.xml").Output()
		if err != nil {
			ctx.StopWithError(500, err)
			return
		}
		wifi_obj := new(WiFiXML)
		if err = xml.Unmarshal(wifi_text, wifi_obj); err != nil {
			ctx.StopWithError(500, err)
			return
		}
		hideSSID := wifi_obj.SoftAp.AP_BOOLS[0].VALUE
		security := wifi_obj.SoftAp.AP_Ints[1].VALUE
		SSIDName := wifi_obj.SoftAp.AP_Strings[0].VALUE
		mac_addr := wifi_obj.SoftAp.AP_Strings[2].VALUE
		password := wifi_obj.SoftAp.AP_Strings[1].VALUE
		bandwidthMode := wifi_obj.SoftAp.AP_BandChannelMap.BandChannel.BandChannel_Ints[0].VALUE
		channel := wifi_obj.SoftAp.AP_BandChannelMap.BandChannel.BandChannel_Ints[1].VALUE
		ctx.JSON(iris.Map{
			"result":        "ok",
			"status":        1,
			"apIsolation":   0,
			"mac_addr":      mac_addr,
			"hideSSID":      hideSSID,
			"SSIDName":      SSIDName,
			"bandwidthMode": bandwidthMode,
			"channel":       channel,
			"security":      security,
			"password":      password,
			// "autoSleep":     0,
		})
	case `ip`:
		clients, err := exec.Command("sh", "-c", "ip -4 neigh | grep ap0 | grep REACHABLE").Output()
		if err != nil {
			ctx.StopWithError(500, err)
			return
		}
		ctx.WriteString(string(clients))
	case `connected_devices`:
		device := iris.Map{
			"index":    "0",
			"hostName": "z",
			"ip_addr":  "192.168.1.100",
			"mac_addr": "d2:36:db:4f:f6:8a",
			"usbShare": "0",
		}
		devices := [...]iris.Map{device, device, device, device}
		ctx.JSON(iris.Map{
			"result":   "ok",
			"totalNum": len(devices),
			"devices":  devices,
		})

	case `flowrate_record`:
		total_send := rand.Int31()
		total_recv := rand.Int31()
		cur_send := rand.Int31()
		cur_recv := rand.Int31()
		// body := fmt.Sprintf(`{	"result": "ok",	"upload": "%v","download": "%v"}`, upload, download)
		// ctx.WriteString(body)
		ctx.JSON(iris.Map{
			"result":     "ok",
			"total_send": total_send,
			"total_recv": total_recv,
			"cur_send":   cur_send,
			"cur_recv":   cur_recv,
		})
	case `navtop_info`:
		ctx.JSON(iris.Map{
			"result":            "ok",
			"batteryRemain":     "46",
			"tobeReadSMS":       "1",
			"language":          "en",
			"totalNumSMS":       "14",
			"isSMSFull":         "0",
			"total_send":        "3450",
			"total_recv":        "3207",
			"cur_send":          "8029",
			"cur_recv":          "5014",
			"threshold_percent": "90",
			"apStatus":          "1",
		})
	case `network_info`:
		networkName, _1 := exec.Command("/system/bin/getprop", "gsm.sim.operator.alpha").Output()
		networkType, _2 := exec.Command("/system/bin/getprop", "gsm.network.type").Output()
		simStatus, _3 := exec.Command("/system/bin/getprop", "gsm.sim.state").Output()
		gprsStatus, _4 := exec.Command("settings", "get", "global", "mobile_data").Output()
		signalStrength, _5 := exec.Command("/system/bin/getprop", "vendor.ril.nw.signalstrength.lte.1").Output()
		if _1 != nil || _2 != nil || _3 != nil || _4 != nil || _5 != nil {
			ctx.StopWithError(500, _1)
			ctx.StopWithError(500, _2)
			ctx.StopWithError(500, _3)
			ctx.StopWithError(500, _4)
			ctx.StopWithError(500, _5)
			return
		}
		ctx.JSON(iris.Map{
			"result":         "ok",
			"networkName":    string(networkName),
			"networkType":    string(networkType),
			"simStatus":      strings.Contains(string(simStatus), "LOADED"),
			"gprsStatus":     string(gprsStatus),
			"signalStrength": strings.Split(string(signalStrength), ",")[0],
		})

	case `network_speed`:
		upload := rand.Int31()
		download := rand.Int31()
		ctx.JSON(iris.Map{
			"result":   "ok",
			"upload":   upload,
			"download": download,
		})
	case `getprop`:
		body, err := exec.Command("/system/bin/getprop").Output()
		if err != nil {
			ctx.StopWithError(500, err)
			return
		}
		ctx.Write(body)
	default:
		ctx.WriteString("REQUEST IS FUCKED BY action = " + action)
	}

	ctx.StatusCode(200)
}

type XmlItems struct {
	XMLName xml.Name `xml:"Envelope"`

	Header struct {
		XMLName xml.Name `xml:"Header"`
		// Autocomplete xml.Attr `xml:"autocomplete,attr"`
		// Valid        xml.Attr `xml:"valid,attr"`
		ID string `xml:"ID"`

		// Text struct {
		// 	XMLName  xml.Name `xml:"text"`
		// 	TextType xml.Attr `xml:"type,attr"`
		// 	Content  string   `xml:",innerxml"`
		// } `xml:"text"`
	} `xml:"Header"`

	Body struct {
		XMLName xml.Name `xml:"Body"`
		Inform  struct {
			XMLName     xml.Name `xml:"Inform"`
			CurrentTime string   `xml:"CurrentTime"`
		} `xml:"Inform"`
	} `xml:"Body"`
}

func create(ctx iris.Context) {
	// var body []byte
	var result XmlItems
	// param := ctx.URLParam("asd")
	// println("param: " + param)
	// println("param: ")
	// if ctx.ReadBody(&body) == nil {
	// 	fmt.Printf("len(body): %v\n", len(body))
	// 	result := XmlEnvelope{}
	// 	xml.Unmarshal(body, &result)
	// 	println("result=" + result.GetIPLocationResult)
	// } else {
	// 	println("FUCK!1")
	// 	return
	// }
	err := ctx.ReadXML(&result)
	if err != nil {
		println(err.Error())
		return
	}
	println("result.Header.ID: " + result.Header.ID)
	println("result.Body.Inform.CurrentTime: " + result.Body.Inform.CurrentTime)
	println("===================================")
	// result := XmlItems{}
	// err := xml.Unmarshal([]byte(query), &result)
	// if err != nil {
	// 	ctx.StopWithError(500, err)
	// 	return
	// }
	// books := []Book{
	// 	{"Mastering Concurrency in Go"},
	// 	{"Go Design Patterns"},
	// 	{"Black Hat Go"},
	// }
	// var xml Xml
	// if ctx.ReadXML(&xml) == nil {
	// 	ctx.Text(xml.Title)
	// 	ctx.StatusCode(200)
	// }
	// println("xml: " + xml.Title)
}
