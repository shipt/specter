package webServer

import (
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/namsral/flag"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shipt/specter/internal/broadcast"
	"github.com/shipt/specter/internal/logprocessor"
	"github.com/shipt/specter/internal/maxmind"
)

// Template contains html/templates
type Template struct {
	templates *template.Template
}

var upgrader = websocket.Upgrader{}
var db string
var mbat string

func init() {
	flag.StringVar(&db, "db", "", "Location of the maxmind DB")
	flag.StringVar(&mbat, "mbat", "", "MapBox Access Token")
}

func wsWrite(broadcaster *broadcast.WebsocketBroadcaster) func(c echo.Context) error {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		broadcaster.Add(ws)
		return nil
	}
}

func logs(lp *logprocessor.Client) func(c echo.Context) error {
	return func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()
		lp.Write(body)
		return nil
	}
}

func version() func(c echo.Context) error {
	return func(c echo.Context) error {
		fileBytes, err := ioutil.ReadFile("/go/src/github.com/shipt/specter/VERSION")
		if err != nil {
			return err
		}

		appVersion := string(fileBytes)
		c.Response().Header().Set("Version", appVersion)
		return c.JSON(204, c.Response)
	}
}

// Render renders the template in Template and returns it.
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func mpapi(c echo.Context) error {
	return c.Render(http.StatusOK, "mpapi", mbat)
}

// Start runs the webserver
func Start() {
	flag.Parse()
	if mbat == "" {
		log.Warn("You must set the mbat flag!")
		log.Fatalf("Flags:\n-db=/location/of/db \n\tLocation of the maxmind DB\n-mbat=TokenString\n\tMapBox Access Token")
	}
	if db == "" {
		log.Warn("You must set the db flag!")
		log.Fatalf("Flags:\n-db=/location/of/db \n\tLocation of the maxmind DB\n-mbat=TokenString\n\tMapBox Access Token")
	}

	t := &Template{
		templates: template.Must(template.ParseGlob("web/index.tmpl")),
	}

	e := echo.New()
	e.Use(middleware.Recover())
	e.Renderer = t
	e.Static("/public", "web/public")

	maxMind, err := maxmind.New(db)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error opening the MaxMind DB")
	}

	lp := logprocessor.New(maxMind)
	broadcaster := broadcast.New(lp.Ingest())
	broadcaster.Broadcast()

	e.GET("/ws", wsWrite(broadcaster))
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error processing on the /ws endpoint")
	}
	e.GET("/", mpapi)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error rendering index.html template")
	}
	e.HEAD("/version", version())
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error processing on the /version endpoint")
	}
	e.POST("/logs", logs(lp))
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Warn("error processing on the /logs endpoint")
	}
	e.Start(":1323")
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("error starting the webserver")
	}
}
