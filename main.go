package main // Look README.md

import (
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"wm24_api/sql"

	"github.com/kataras/iris/v12"
)

func main() {

	app := iris.Default()
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
