package main // Look README.md

import (
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"wm24_api/sql"

	"github.com/kataras/iris/v12"
)

func main() {

	app := iris.Default()
	tmpl := iris.HTML("./", ".html")
	app.RegisterView(tmpl)
	app.HandleDir("./", http.Dir("./"))
	// db, secret := init_db()
	// subRouter := api.Router(db, secret)
	// app.PartyFunc("/", subRouter)

	/*************************Custom Routers****************************/

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

func run_action(ctx iris.Context) {

	action := ctx.Params().Get("action")
	switch action {
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
			"simStatus":      string(simStatus),
			"gprsStatus":     string(gprsStatus),
			"signalStrength": string(signalStrength),
		})
	case `network_speed`:
		upload := rand.Int31()
		download := rand.Int31()
		body := fmt.Sprintf(`{	"result": "ok",	"upload": "%v","download": "%v"}`, upload, download)
		ctx.WriteString(body)
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
	Header  struct {
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
