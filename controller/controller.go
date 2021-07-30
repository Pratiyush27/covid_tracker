package controller

import (
	"context"
	"fmt"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"log"
	"github.com/labstack/echo/v4"
	"strings"
	"encoding/json"
	"InshortsAssignment/models"
	"InshortsAssignment/cache"
	"github.com/go-redis/redis/v8"
	"time"
	// "crypto/tls" https://api.covid19india.org/csv/latest/state_wise.csv
)

var api_key = "rhaqcJgeUGWkp4-FIWHcW1oDi89-XCtBssg5Nzdy68Y"


func process(s string) (result string) {
	flag := 0
	for _, v := range s {
      		if(v=='"') {
			flag = (flag+1)%2
		} else {
			if(flag==0) {
				result = result + string(v)
			}
		}
			
   	}
	return 
}





func GetIssuesByCode(code string, myCache *redis.Client) models.Entry {
	filter := bson.D{{Key: "state", Value: code}}
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	 if err != nil {
	  log.Fatal(err)
	 }
	 err = client.Connect(context.Background())
	 if err != nil {
	  log.Fatal(err)
	 }
	db := client.Database("SAMPLETRIAL")
	cur,err := db.Collection("STATEWISE").Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	var elem models.Entry
	// Get the next result from the cursor
	for cur.Next(context.Background()) {
		err := cur.Decode(&elem)
		if err != nil {
		log.Fatal(err)
		}
		return elem
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.Background())

	// Add to Cache
	err = cache.Set(myCache,code, elem, 30*time.Minute)
	if err != nil {
		log.Fatal(err)
	}	

	return elem
}

func Getstatefromlatilongi(c echo.Context) error {
	myCache := cache.InitRedisCache()
	latitudeval := c.FormValue("latitude")
	longitudeval := c.FormValue("longitude")


	baseUrl := "https://revgeocode.search.hereapi.com/v1/revgeocode?apiKey="
	req := "&at=" + "" +string(latitudeval)+","+string(longitudeval)+""
	final_req := baseUrl + api_key + req
	//log.Printf(final_req)
	resp, err := http.Get(final_req)
	if err != nil {
		log.Fatalln(err)
		return c.String(http.StatusOK, "lati:" + latitudeval + ", longi:" + longitudeval+" ")
	}
	body, err := ioutil.ReadAll(resp.Body)

	var myState = ""
	var myCountry = ""
	var data models.Output
	json.Unmarshal(body, &data)
	// fmt.Println(data)
	for _, add := range data.Items {
		myState = add.Address.State
		myCountry = add.Address.CountryName
	}
	 
	if(strings.Compare(myCountry, "India")==0)	{
		log.Printf(myState)
				
		// Check Cache

		res,err := cache.Get(myCache,myState)
		if err != nil {
			log.Fatal(err)
		}

		if res != nil {
			// cache exist
			var result models.Entry

			err := json.Unmarshal(res, &result)
			if err != nil {
				log.Fatal(err)
			}

			return c.String(http.StatusOK, "State is "+result.State+ ", Active Cases are "+result.Cases+", Last Updated Time is "+result.Last_Updated+" Hit from Cache ")	
		}

		ans:= GetIssuesByCode(myState,myCache)
		total := GetIssuesByCode("Total",myCache)
		return c.String(http.StatusOK, "State is "+ans.State+ ", Active Cases are "+ans.Cases+", Last Updated Time is "+ans.Last_Updated + ",  Total Active Cases in India are "+total.Cases+", Last Updated Time for Total Indian Cases is "+total.Last_Updated)	
	} else	{
		log.Printf("error")
	}

	bodyString := string(body)
	sb := string(bodyString)
	log.Printf(sb)
	log.Printf("\n")
	log.Printf(final_req)
	
	
	return c.String(http.StatusOK, "lati:" + latitudeval + ", longi:" + longitudeval+" ")
	
}

func Updatemongodb(c echo.Context) error {
	myCache := cache.InitRedisCache()
	err1 := cache.FlushDB(myCache)
	if err1 != nil {
		log.Fatal(err1)
	}	
	response, err := http.Get("https://api.covid19india.org/csv/latest/state_wise.csv")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	dataset := string(responseData)
	dataset = process(dataset)
	rows := strings.Split(dataset,"\n")

	


	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	 if err != nil {
	  log.Fatal(err)
	 }
	 err = client.Connect(context.Background())
	 if err != nil {
	  log.Fatal(err)
	 }
	 db := client.Database("SAMPLETRIAL")

	 // Insert people into DB
	 var ppl []interface{}
	 
	
	

	for _, row := range rows {
      		cols := strings.Split(row,",")
		fmt.Println(cols[0] + " "+cols[4]+" "+cols[5])
		ppl = append(ppl, models.Entry{cols[0],cols[4],cols[5]})
   	}

	_, err = db.Collection("STATEWISE").DeleteMany(context.TODO(), bson.D{{}})
	 if err != nil {
	  log.Fatal(err)
	 }

	_, err = db.Collection("STATEWISE").InsertMany(context.Background(), ppl)
	 if err != nil {
	  log.Fatal(err)
	 }
	return c.String(http.StatusOK, "DB UPDATED ")
	
}
