package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

// Global Variables
var db *sqlx.DB

// SQL Create Tables
const create_assets = `CREATE TABLE IF NOT EXISTS ASSETS (
	ASSET_ID VARCHAR(100) NOT NULL,
	ASSET_NAME VARCHAR(100) NOT NULL,
	ASSET_ZONE VARCHAR(100) NOT NULL,
	DESCRIPTION TEXT
  );`

const create_connections = `CREATE TABLE IF NOT EXISTS CONNS (
	CONN_ID VARCHAR(100) PRIMARY KEY,
	SOURCE_ASSET_ID VARCHAR(100), 
	DEST_ASSET_ID VARCHAR(100), 
	PROTOCOL VARCHAR(100),
	ENCRYPTION VARCHAR(100),
	SERVER_AUTHENTICATION VARCHAR(100),
	CLIENT_AUTHENTICATION VARCHAR(100),
	CLIENT_AUTHORISATION VARCHAR(100),
	SERVER_CRL VARCHAR(100),
	CLIENT_CRL VARCHAR(100),
	DESCRIPTION TEXT
  );`

const assetTest1 = "INSERT INTO ASSETS (ASSET_ID,ASSET_NAME,ASSET_ZONE,DESCRIPTION) VALUES ('1234','Test  1','Zone  1','a')"
const assetTest2 = "INSERT INTO ASSETS (ASSET_ID,ASSET_NAME,ASSET_ZONE,DESCRIPTION) VALUES ('5678','Test  2','Zone  2','TestAsset  2')"
const connTest1 = "INSERT INTO CONNS(CONN_ID,SOURCE_ASSET_ID,DEST_ASSET_ID,PROTOCOL,ENCRYPTION,SERVER_AUTHENTICATION,CLIENT_AUTHENTICATION,CLIENT_AUTHORISATION,SERVER_CRL,CLIENT_CRL,DESCRIPTION) VALUES ('9999','1234','5678','a','b','c','d','e','f','g','h')"
const connTest2 = "INSERT INTO CONNS(CONN_ID,SOURCE_ASSET_ID,DEST_ASSET_ID,PROTOCOL,ENCRYPTION,SERVER_AUTHENTICATION,CLIENT_AUTHENTICATION,CLIENT_AUTHORISATION,SERVER_CRL,CLIENT_CRL,DESCRIPTION) VALUES ('9998','1234','5678','9','8','7','6','5','4','3','2')"

// Structs
type assets struct {
	ASSET_ID    string `db:"ASSET_ID"`
	ASSET_NAME  string `db:"ASSET_NAME"`
	ASSET_ZONE  string `db:"ASSET_ZONE"`
	DESCRIPTION string `db:"DESCRIPTION"`
}

type connections struct {
	CONN_ID               string `db:"CONN_ID"`
	SOURCE_ASSET_ID       string `db:"SOURCE_ASSET_ID"`
	DEST_ASSET_ID         string `db:"DEST_ASSET_ID"`
	PROTOCOL              string `db:"PROTOCOL"`
	ENCRYPTION            string `db:"ENCRYPTION"`
	SERVER_AUTHENTICATION string `db:"SERVER_AUTHENTICATION"`
	CLIENT_AUTHENTICATION string `db:"CLIENT_AUTHENTICATION"`
	CLIENT_AUTHORISATION  string `db:"CLIENT_AUTHORISATION"`
	SERVER_CRL            string `db:"SERVER_CRL"`
	CLIENT_CRL            string `db:"CLIENT_CRL"`
	DESCRIPTION           string `db:"DESCRIPTION"`
}

type connectionlist struct {
	CONN_ID               string `db:"CONN_ID"`
	SOURCE_ASSET_ID       string `db:"SOURCE_ASSET_ID"`
	DEST_ASSET_ID         string `db:"DEST_ASSET_ID"`
	PROTOCOL              string `db:"PROTOCOL"`
	ENCRYPTION            string `db:"ENCRYPTION"`
	SERVER_AUTHENTICATION string `db:"SERVER_AUTHENTICATION"`
	CLIENT_AUTHENTICATION string `db:"CLIENT_AUTHENTICATION"`
	CLIENT_AUTHORISATION  string `db:"CLIENT_AUTHORISATION"`
	SERVER_CRL            string `db:"SERVER_CRL"`
	CLIENT_CRL            string `db:"CLIENT_CRL"`
	DESCRIPTION           string `db:"DESCRIPTION"`
	SOURCE_ASSET_NAME     string `db:"SOURCE_ASSET_NAME"`
	SOURCE_ASSET_ZONE     string `db:"SOURCE_ASSET_ZONE"`
	DEST_ASSET_NAME       string `db:"DEST_ASSET_NAME"`
	DEST_ASSET_ZONE       string `db:"DEST_ASSET_ZONE"`
}

type zones struct {
	ASSET      string `db:"ASSET"`
	ASSET_ZONE string `db:"ASSET_ZONE"`
}

type diagconn struct {
	SOURCE_ASSET_NAME string `db:"SOURCE_ASSET_NAME"`
	CONN_ID           string `db:"CONN_ID"`
	PROTOCOL          string `db:"PROTOCOL"`
	ENCRYPTION        string `db:"ENCRYPTION"`
	DEST_ASSET_NAME   string `db:"DEST_ASSET_NAME"`
}

// Main
func main() {
	//Start DB
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	createTables(db)
	createTestData(db)

	// Web Server Config
	r := gin.Default()
	r.Use(serverHeader)
	r.LoadHTMLGlob("views/**/*")

	r.GET("/", func(c *gin.Context) {
		assetsData := []assets{}
		err = db.Select(&assetsData, "SELECT * FROM ASSETS")

		// de duplicate a lot of this mess
		const connSql = `SELECT c.CONN_ID AS 'CONN_ID', c.SOURCE_ASSET_ID AS 'SOURCE_ASSET_ID',
		c.DEST_ASSET_ID AS 'DEST_ASSET_ID',c.PROTOCOL AS 'PROTOCOL',c.ENCRYPTION AS 'ENCRYPTION',c.SERVER_AUTHENTICATION AS 'SERVER_AUTHENTICATION',
		c.CLIENT_AUTHENTICATION AS 'CLIENT_AUTHENTICATION',c.CLIENT_AUTHORISATION AS 'CLIENT_AUTHORISATION',c.SERVER_CRL AS 'SERVER_CRL',
		c.CLIENT_CRL AS 'CLIENT_CRL',c.DESCRIPTION AS 'DESCRIPTION',s.ASSET_NAME as 'SOURCE_ASSET_NAME',
		s.ASSET_ZONE as 'SOURCE_ASSET_ZONE',d.ASSET_NAME as 'DEST_ASSET_NAME',
		d.ASSET_ZONE as 'DEST_ASSET_ZONE' from CONNS c 
		left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID
		left join ASSETS d on c.DEST_ASSET_ID = d.ASSET_ID
		`
		const appZone = `SELECT replace(s.ASSET_NAME," ","_") as 'ASSET',
		replace(s.ASSET_ZONE," ","_") as 'ASSET_ZONE' 
					from CONNS c left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID
                    UNION SELECT replace(s.ASSET_NAME," ","_") as 'ASSET',
					replace(s.ASSET_ZONE," ","_") as 'ASSET_ZONE'
					 from CONNS c left join ASSETS s on c.DEST_ASSET_ID = s.ASSET_ID
                    `
		const diagList = `SELECT c.CONN_ID AS 'CONN_ID', replace(c.PROTOCOL," ","_") AS 'PROTOCOL',replace(c.ENCRYPTION," ","_") AS 'ENCRYPTION',replace(s.ASSET_NAME," ","_") as 'SOURCE_ASSET_NAME',
		replace(d.ASSET_NAME," ","_") as 'DEST_ASSET_NAME' from CONNS c left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID left join ASSETS d on c.DEST_ASSET_ID = d.ASSET_ID
		`

		connectionsData := []connectionlist{}
		err = db.Select(&connectionsData, connSql)

		zonesData := []zones{}
		err = db.Select(&zonesData, appZone)

		diagData := []diagconn{}
		err = db.Select(&diagData, diagList)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"assetsData": assetsData, "connectionsData": connectionsData, "zonesData": zonesData, "diagData": diagData,
		})
	})

	r.GET("/mapNew", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mapnew.html", gin.H{})
	})

	r.GET("/mapLoad", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mapload.html", gin.H{})
	})

	r.GET("/mapSave", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mapsave.html", gin.H{})
	})

	r.GET("/assetAdd", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assetadd.html", gin.H{})
	})

	r.GET("/assetDelete/:id", func(c *gin.Context) {
		ids := c.Param("id")
		assetsData := assets{}
		err = db.Get(&assetsData, "SELECT * FROM ASSETS WHERE ASSET_ID=?", ids)
		c.HTML(http.StatusOK, "assetdelete.html", gin.H{
			"assetsData": assetsData,
		})
	})

	r.GET("/assetEdit/:id", func(c *gin.Context) {
		ids := c.Param("id")
		assetsData := assets{}
		err = db.Get(&assetsData, "SELECT * FROM ASSETS WHERE ASSET_ID=?", ids)
		c.HTML(http.StatusOK, "assetedit.html", gin.H{
			"assetsData": assetsData,
		})
	})

	r.GET("/connectionAdd", func(c *gin.Context) {
		assetsData := []assets{}
		err = db.Select(&assetsData, "SELECT * FROM ASSETS")
		c.HTML(http.StatusOK, "connectionadd.html", gin.H{
			"assetsData": assetsData,
		})
	})

	r.GET("/connectionEdit/:id", func(c *gin.Context) {
		ids := c.Param("id")
		const connSql = `SELECT c.CONN_ID AS 'CONN_ID', c.SOURCE_ASSET_ID AS 'SOURCE_ASSET_ID',
		c.DEST_ASSET_ID AS 'DEST_ASSET_ID',c.PROTOCOL AS 'PROTOCOL',c.ENCRYPTION AS 'ENCRYPTION',c.SERVER_AUTHENTICATION AS 'SERVER_AUTHENTICATION',
		c.CLIENT_AUTHENTICATION AS 'CLIENT_AUTHENTICATION',c.CLIENT_AUTHORISATION AS 'CLIENT_AUTHORISATION',c.SERVER_CRL AS 'SERVER_CRL',
		c.CLIENT_CRL AS 'CLIENT_CRL',c.DESCRIPTION AS 'DESCRIPTION',s.ASSET_NAME as 'SOURCE_ASSET_NAME',
		s.ASSET_ZONE as 'SOURCE_ASSET_ZONE',d.ASSET_NAME as 'DEST_ASSET_NAME',
		d.ASSET_ZONE as 'DEST_ASSET_ZONE' from CONNS c 
		left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID
		left join ASSETS d on c.DEST_ASSET_ID = d.ASSET_ID where c.CONN_ID=?`

		assetsData := []assets{}
		err = db.Select(&assetsData, "SELECT * FROM ASSETS")

		connectionsData := connectionlist{}
		err = db.Get(&connectionsData, connSql, ids)

		c.HTML(http.StatusOK, "connectionedit.html", gin.H{
			"assetsData": assetsData, "connectionsData": connectionsData,
		})
	})

	r.GET("/connectionClone/:id", func(c *gin.Context) {
		ids := c.Param("id")
		const connSql = `SELECT c.CONN_ID AS 'CONN_ID', c.SOURCE_ASSET_ID AS 'SOURCE_ASSET_ID',
		c.DEST_ASSET_ID AS 'DEST_ASSET_ID',c.PROTOCOL AS 'PROTOCOL',c.ENCRYPTION AS 'ENCRYPTION',c.SERVER_AUTHENTICATION AS 'SERVER_AUTHENTICATION',
		c.CLIENT_AUTHENTICATION AS 'CLIENT_AUTHENTICATION',c.CLIENT_AUTHORISATION AS 'CLIENT_AUTHORISATION',c.SERVER_CRL AS 'SERVER_CRL',
		c.CLIENT_CRL AS 'CLIENT_CRL',c.DESCRIPTION AS 'DESCRIPTION',s.ASSET_NAME as 'SOURCE_ASSET_NAME',
		s.ASSET_ZONE as 'SOURCE_ASSET_ZONE',d.ASSET_NAME as 'DEST_ASSET_NAME',
		d.ASSET_ZONE as 'DEST_ASSET_ZONE' from CONNS c 
		left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID
		left join ASSETS d on c.DEST_ASSET_ID = d.ASSET_ID where c.CONN_ID=?`

		assetsData := []assets{}
		err = db.Select(&assetsData, "SELECT * FROM ASSETS")

		connectionsData := connectionlist{}
		err = db.Get(&connectionsData, connSql, ids)

		c.HTML(http.StatusOK, "connectionclone.html", gin.H{
			"assetsData": assetsData, "connectionsData": connectionsData,
		})
	})

	r.GET("/connectionDelete/:id", func(c *gin.Context) {
		ids := c.Param("id")
		const connSql = `SELECT c.CONN_ID AS 'CONN_ID', c.SOURCE_ASSET_ID AS 'SOURCE_ASSET_ID',
		c.DEST_ASSET_ID AS 'DEST_ASSET_ID',c.PROTOCOL AS 'PROTOCOL',c.ENCRYPTION AS 'ENCRYPTION',c.SERVER_AUTHENTICATION AS 'SERVER_AUTHENTICATION',
		c.CLIENT_AUTHENTICATION AS 'CLIENT_AUTHENTICATION',c.CLIENT_AUTHORISATION AS 'CLIENT_AUTHORISATION',c.SERVER_CRL AS 'SERVER_CRL',
		c.CLIENT_CRL AS 'CLIENT_CRL',c.DESCRIPTION AS 'DESCRIPTION',s.ASSET_NAME as 'SOURCE_ASSET_NAME',
		s.ASSET_ZONE as 'SOURCE_ASSET_ZONE',d.ASSET_NAME as 'DEST_ASSET_NAME',
		d.ASSET_ZONE as 'DEST_ASSET_ZONE' from CONNS c 
		left join ASSETS s on c.SOURCE_ASSET_ID = s.ASSET_ID
		left join ASSETS d on c.DEST_ASSET_ID = d.ASSET_ID where c.CONN_ID=?`

		connectionsData := connectionlist{}
		err = db.Get(&connectionsData, connSql, ids)

		c.HTML(http.StatusOK, "connectiondelete.html", gin.H{
			"connectionsData": connectionsData,
		})
	})

	////////////////// POSTS N STUFF

	r.POST("/mapNew", func(c *gin.Context) {
		const sql = "DELETE FROM ASSETS"
		const sqlD = "DELETE FROM CONNS"
		_, err = db.Exec(sql)
		_, err = db.Exec(sqlD)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/connectionDelete/:id", func(c *gin.Context) {
		ids := c.Param("id")
		const sql = "DELETE FROM CONNS WHERE CONN_ID=?"

		_, err = db.Exec(sql, ids)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/assetDelete/:id", func(c *gin.Context) {
		ids := c.Param("id")
		const sql = "DELETE FROM ASSETS WHERE ASSET_ID=?"
		const sqlD = "DELETE FROM CONNS WHERE SOURCE_ASSET_ID=? OR DEST_ASSET_ID =?"

		_, err = db.Exec(sql, ids)
		_, err = db.Exec(sqlD, ids, ids)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/assetAdd/", func(c *gin.Context) {
		//an := strings.ReplaceAll(c.Request.FormValue("assetName"), " ", "_")
		an := c.Request.FormValue("assetName")
		zn := c.Request.FormValue("zoneName")
		de := c.Request.FormValue("description")

		sql := "INSERT INTO ASSETS (ASSET_ID,ASSET_NAME,ASSET_ZONE,DESCRIPTION) VALUES (?,?,?,?)"
		_, err = db.Exec(sql, uuid.NewString(), an, zn, de)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/connectionAdd/", func(c *gin.Context) {
		//an := strings.ReplaceAll(c.Request.FormValue("assetName"), " ", "_")
		sa := c.Request.FormValue("sourceApp")
		da := c.Request.FormValue("destApp")
		pr := c.Request.FormValue("proto")
		en := c.Request.FormValue("enc")
		sva := c.Request.FormValue("serverAuth")
		cla := c.Request.FormValue("clientAuth")
		clz := c.Request.FormValue("clientAuthz")
		svc := c.Request.FormValue("serverCRL")
		clc := c.Request.FormValue("clientCRL")
		de := c.Request.FormValue("desc")

		sql := "INSERT INTO CONNS (CONN_ID,SOURCE_ASSET_ID,DEST_ASSET_ID,PROTOCOL,ENCRYPTION,SERVER_AUTHENTICATION,CLIENT_AUTHENTICATION,CLIENT_AUTHORISATION,SERVER_CRL,CLIENT_CRL,DESCRIPTION) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
		_, err = db.Exec(sql, uuid.NewString(), sa, da, pr, en, sva, cla, clz, svc, clc, de)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/connectionClone/", func(c *gin.Context) {
		//an := strings.ReplaceAll(c.Request.FormValue("assetName"), " ", "_")
		sa := c.Request.FormValue("sourceApp")
		da := c.Request.FormValue("destApp")
		pr := c.Request.FormValue("proto")
		en := c.Request.FormValue("enc")
		sva := c.Request.FormValue("serverAuth")
		cla := c.Request.FormValue("clientAuth")
		clz := c.Request.FormValue("clientAuthz")
		svc := c.Request.FormValue("serverCRL")
		clc := c.Request.FormValue("clientCRL")
		de := c.Request.FormValue("desc")

		sql := "INSERT INTO CONNS (CONN_ID,SOURCE_ASSET_ID,DEST_ASSET_ID,PROTOCOL,ENCRYPTION,SERVER_AUTHENTICATION,CLIENT_AUTHENTICATION,CLIENT_AUTHORISATION,SERVER_CRL,CLIENT_CRL,DESCRIPTION) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
		_, err = db.Exec(sql, uuid.NewString(), sa, da, pr, en, sva, cla, clz, svc, clc, de)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/connectionEdit/:id", func(c *gin.Context) {
		sa := c.Request.FormValue("sourceApp") + ""
		da := c.Request.FormValue("destApp") + ""
		pr := c.Request.FormValue("proto") + ""
		en := c.Request.FormValue("enc") + ""
		sva := c.Request.FormValue("serverAuth") + ""
		cla := c.Request.FormValue("clientAuth") + ""
		clz := c.Request.FormValue("clientAuthz") + ""
		svc := c.Request.FormValue("serverCRL") + ""
		clc := c.Request.FormValue("clientCRL") + ""
		de := c.Request.FormValue("desc") + ""
		ids := c.Param("id") + ""

		const sql = "UPDATE CONNS SET SOURCE_ASSET_ID=?,DEST_ASSET_ID=?,PROTOCOL=?,ENCRYPTION=?,SERVER_AUTHENTICATION=?,CLIENT_AUTHENTICATION=?,CLIENT_AUTHORISATION=?,SERVER_CRL=?,CLIENT_CRL=?,DESCRIPTION=? WHERE CONN_ID=?"
		_, err = db.Exec(sql, sa, da, pr, en, sva, cla, clz, svc, clc, de, ids)
		c.Redirect(http.StatusFound, "/")
	})

	r.POST("/assetEdit/:id", func(c *gin.Context) {
		an := c.Request.FormValue("assetName") + ""
		az := c.Request.FormValue("zoneName") + ""
		de := c.Request.FormValue("description") + ""
		ids := c.Param("id") + ""

		const sql = "UPDATE ASSETS SET ASSET_NAME=?,ASSET_ZONE=?,DESCRIPTION=? WHERE ASSET_ID=?"
		_, err = db.Exec(sql, an, az, de, ids)
		c.Redirect(http.StatusFound, "/")
	})

	r.Run(":80") // listen and serve on 0.0.0.0:8080
}

// Server Headers for Security
func serverHeader(c *gin.Context) {
	c.Header("Server", "Some-Play-Server")
	c.Header("X-Powered-By", "Yo Momma")
}

func createTables(db *sqlx.DB) {
	var err error
	_, err = db.Exec(create_assets)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(create_connections)
	if err != nil {
		log.Fatal(err)
	}
}

func createTestData(db *sqlx.DB) {
	var err error
	_, err = db.Exec(assetTest1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(assetTest2)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(connTest1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(connTest2)
	if err != nil {
		log.Fatal(err)
	}

}
