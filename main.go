package main // Look README.md

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
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
		api.Get("/{action}", dispatcher)
		api.Post("/{action}", dispatcher)
	}

	// for static html files
	html := app.Party("/")
	{
		html.Use(iris.Compression)
		html.Get("/{page}", func(ctx iris.Context) {
			page := ctx.Params().Get("page")
			ctx.View(page)
		})
		// html.Post("/{action}", dispatcher)
	}

	/*************************Starting Server****************************/
	addr := fmt.Sprintf("0.0.0.0:%s", getenv("PORT", "9001"))
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

func dispatcher(ctx iris.Context) {

	action := ctx.Params().Get("action")
	switch action {

	case `get_device_info`:
		firmwarewVersion, _ := exec.Command("/system/bin/getprop", "ro.mediatek.version.release").Output()
		serialNumber, _ := exec.Command("/system/bin/getprop", "ro.serialno").Output()
		imei, _ := exec.Command("sh", "-c", "cmd phone get-imei 0").Output()
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

		wanIP_text, _ := exec.Command("sh", "-c", "(ifconfig ccmni0 && ifconfig ccmni1) | grep 'inet addr:'").Output()

		wanIP := ""
		if len(wanIP_text) > 1 {
			wanIP = strings.ReplaceAll(
				strings.ReplaceAll(valFilter(wanIP_text), "inet addr:", ""),
				"  Mask:255.0.0.0", "")
		}

		ctx.JSON(iris.Map{
			"result":           "ok",
			"serialNumber":     valFilter(serialNumber),
			"imei":             valFilter(imei),
			"imsi":             "00000000000000000",
			"hardwareVersion":  "1.0.0",
			"softwarewVersion": "随便自定义1_1_1 和WEBUI重复了？",
			"firmwarewVersion": valFilter(firmwarewVersion),
			"webUIVersion":     "随便自定义1_1_1",
			"mac":              mac_addr,
			"wanIP":            wanIP,
		})
	case `get_pin_setting`:
		pinRemain, err1 := exec.Command("/system/bin/getprop", "vendor.gsm.sim.retry.pin1").Output()
		pinEnabled, err2 := exec.Command("/system/bin/getprop", "gsm.slot1.num.pin1").Output()
		pinStatus, err3 := exec.Command("/system/bin/getprop", "gsm.slot1.num.pin1").Output()
		_pinRemain, _ := strconv.Atoi(valFilter(pinRemain[:len(pinRemain)-1]))
		_pinEnabled, _ := strconv.Atoi(valFilter(pinEnabled[:len(pinEnabled)-1]))
		_pinStatus, _ := strconv.Atoi(valFilter(pinStatus[:len(pinStatus)-1]))
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
		ctx.WriteString(valFilter(clients))
	case `connected_devices`:
		//ip neigh show dev ap0
		clients_buf, _ := exec.Command("sh", "-c", "ip -4 neigh show dev ap0").Output()
		devices := []iris.Map{}
		clients_list := strings.Split(valFilter(clients_buf), "\n")

		for i, v := range clients_list {
			client_map := strings.Split(v, " ")
			if len(client_map) > 2 {
				device := iris.Map{
					"index":    i,
					"hostName": client_map[2],
					"ip_addr":  client_map[0],
					"mac_addr": client_map[2],
					"usbShare": "0",
				}
				devices = append(devices, device)
			}
		}

		ctx.JSON(iris.Map{
			"result":   "ok",
			"totalNum": len(devices),
			"devices":  devices,
		})
	case `get_web_language`:
		config_buf, _ := os.ReadFile("config.json")
		config := iris.Map{}
		_ = json.Unmarshal(config_buf, &config)
		ctx.JSON(iris.Map{
			"result":   "ok",
			"message":  config["language"],
			"language": config["language"],
		})
	case `set_web_language`:
		params := PostJsonDecoder(ctx, `set_web_language`)

		config_buf, _ := os.ReadFile("config.json")
		config := iris.Map{}
		_ = json.Unmarshal(config_buf, &config)
		config["language"] = params["set_web_language"]
		config_buf, _ = json.Marshal(config)
		os.WriteFile("config.json", config_buf, 0666)

		ctx.JSON(iris.Map{
			"result":  "ok",
			"message": params["set_web_language"],
		})
	case `flowrate_record`:
		total_send := rand.Int31()
		total_recv := rand.Int31()
		cur_send := rand.Int31()
		cur_recv := rand.Int31()
		// body := fmt.Sprintf(`{	"result": "ok",	"upload": "%v","download": "%v"}`, upload, download)
		// ctx.WritevalFilter(body)
		ctx.JSON(iris.Map{
			"result":     "ok",
			"total_send": total_send,
			"total_recv": total_recv,
			"cur_send":   cur_send,
			"cur_recv":   cur_recv,
		})
	case `navtop_info`:
		batteryRemain, _ := exec.Command("sh", "-c", "dumpsys battery get level").Output()
		apStatus, _ := exec.Command("sh", "-c", "ifconfig ap0 | grep RUNNING").Output()

		ctx.JSON(iris.Map{
			"result":            "ok",
			"batteryRemain":     valFilter(batteryRemain),
			"language":          "en",
			"tobeReadSMS":       "1",
			"totalNumSMS":       "14",
			"isSMSFull":         "0",
			"total_send":        "3450",
			"total_recv":        "3207",
			"cur_send":          "8029",
			"cur_recv":          "5014",
			"threshold_percent": "90",
			"apStatus":          strings.Contains(valFilter(apStatus), "RUNNING"),
		})
	case `set_network_config`:
		params := PostJsonDecoder(ctx, `set_network_config`)

		// 11 = 4G
		// 9 = 4G/3G
		// 3 = 3G
		switch params["networkMode"] {
		case "0":
			params["networkMode"] = "9"
		case "1":
			params["networkMode"] = "11"
		case "2":
			params["networkMode"] = "3"
		default:
			ctx.StopWithText(500, "param error")
			return
		}

		gprsStatus := fmt.Sprintf("settings put global mobile_data %v", params["gprsStatus"])
		roamingStatus := fmt.Sprintf("settings put global data_roaming %v", params["roamingStatus"])
		networkType := fmt.Sprintf("settings put global preferred_network_mode %v", params["networkMode"])
		_, _ = exec.Command("sh", "-c", gprsStatus).Output()
		_, _ = exec.Command("sh", "-c", networkType).Output()
		_, _ = exec.Command("sh", "-c", roamingStatus).Output()
		ctx.JSON(iris.Map{
			"result": "ok",
		})
	case `network_setting`:
		roamingStatus, _ := exec.Command("sh", "-c", "settings get global data_roaming").Output()
		networkType, _ := exec.Command("sh", "-c", "settings get global preferred_network_mode").Output()
		gprsStatus, _ := exec.Command("settings", "get", "global", "mobile_data").Output()
		_networkType := ""
		switch valFilter(networkType) {
		case "9":
			_networkType = "0"
		case "11":
			_networkType = "1"
		case "3":
			_networkType = "2"
		case "33,33":
			_networkType = "33"
		default:
			_networkType = valFilter(networkType)
			// ctx.StopWithText(500, "param error"+valFilter(networkType))
			// return
		}
		ctx.JSON(iris.Map{
			"result":        "ok",
			"gprsStatus":    valFilter(gprsStatus),
			"roamingStatus": valFilter(roamingStatus),
			"networkMode":   _networkType,
		})
	case `network_info`:
		networkName, _ := exec.Command("/system/bin/getprop", "gsm.sim.operator.alpha").Output()
		networkType, _ := exec.Command("/system/bin/getprop", "gsm.network.type").Output()
		simStatus, _ := exec.Command("/system/bin/getprop", "gsm.sim.state").Output()
		gprsStatus, _ := exec.Command("settings", "get", "global", "mobile_data").Output()
		signalStrength, _ := exec.Command("/system/bin/getprop", "vendor.ril.nw.signalstrength.lte.1").Output()
		ctx.JSON(iris.Map{
			"result":         "ok",
			"networkName":    valFilter(networkName),
			"networkType":    valFilter(networkType),
			"simStatus":      strings.Contains(valFilter(simStatus), "LOADED"),
			"gprsStatus":     valFilter(gprsStatus),
			"signalStrength": strings.Split(valFilter(signalStrength), ",")[0],
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
	case `restart`:
		params := PostJsonDecoder(ctx, `restart`)
		if params["restart"] == "1" {
			//async
			go exec.Command("sh", "-c", "sleep 5 && reboot").Output()
		}
		ctx.JSON(iris.Map{
			"result": "ok",
			"params": params["restart"],
		})
	default:
		ctx.WriteString("REQUEST IS FAILED BY action = " + action)
	}

	ctx.StatusCode(200)
}

func PostJsonDecoder(ctx iris.Context, action string) map[string]interface{} {
	temp := make(iris.Map)
	var body_buffer []byte
	body_buffer, _ = ctx.GetBody()
	values, _ := url.ParseQuery(valFilter(body_buffer))

	err := json.Unmarshal([]byte(values.Get(action)), &temp)
	if err != nil {
		// this is only for int value but not jsons
		// if values.Get(action) is not like {blablabla}
		return iris.Map{
			action: values.Get(action),
		}
	}
	return temp
}

func valFilter(val []byte) string {
	return strings.TrimRight(string(val), "\n")
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
