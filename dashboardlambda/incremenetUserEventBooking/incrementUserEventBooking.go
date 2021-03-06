package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

// AWS lambda request
type MyRequest struct {
	Bucket  string `json:"bucket"`
	Key     string `json:"user_uuid"` // The organizer of the event
	EventID string `json:"eventId"`   // Organizer of the event

}

// AWS lambda response
type MyResponse struct {
	Status bool `json:"status"`
}

// RIAK Network Load Balancer Reesponse

type Dashboard struct {
	PostedEvents []PostedEvent `json:"postedEvents"`
	BookedEvents []BookedEvent `json:"bookedEvents"`
}

type PostedEvent struct {
	OrgID            string `json:"orgId"`
	EventName        string `json:"eventName"`
	Location         string `json:"location"`
	Date             string `json:"date"`
	NumberOfViews    int    `json:"numberOfviews"`
	NumberOfBookings int    `json:"numberOfBookings"`
}

type BookedEvent struct {
	OrgID         string `json:"orgId"`
	EventName     string `json:"eventName"`
	Date          string `json:"date"`
	TimeOfBooking string `json:"timeOfBooking"`
	Location      string `json:"location"`
}

func incrementUserEventBookings(request MyRequest) (MyResponse, error) {

	var nlb = os.Getenv("riak_cluster_nlb")
	var port = os.Getenv("nlb_port")

	// Unboxing Request
	var bucket = request.Bucket
	var key = request.Key
	//var eventID = request.EventID

	// Getting the details of the user from RIAK
	fmt.Println("Arihant")
	var url = fmt.Sprintf("http://%s:%s/buckets/%s/keys/%s", nlb, port, bucket, key)
	fmt.Println(url)
	responseUserDetails, err := http.Get(url)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
		return MyResponse{Status: false}, err
	}
	defer responseUserDetails.Body.Close()
	responseUserDetailsBody, _ := ioutil.ReadAll(responseUserDetails.Body)
	var dashboard Dashboard
	errResponseUserDetailsBody := json.Unmarshal(responseUserDetailsBody, &dashboard)
	if errResponseUserDetailsBody != nil {
		log.Fatalf("ERROR: %s", errResponseUserDetailsBody)
		return MyResponse{Status: false}, errResponseUserDetailsBody
	}

	dashboardString, _ := json.Marshal(dashboard)
	fmt.Println(dashboardString)
	/*
		// Find the particular event
		for i := range dashboard.PostedEvents {
			if dashboard.PostedEvents[i].OrgID == eventID {
				views := dashboard.PostedEvents[i].NumberOfBookings
				dashboard.PostedEvents[i].NumberOfBookings = views + 1
				break
			}
		}

		dashboardString, _ := json.Marshal(dashboard)
		payload := bytes.NewBuffer(dashboardString)
		req, _ := http.NewRequest("PUT", url+"?returnbody=true", payload)
		req.Header.Add("Content-Type", "application/json")
		responsePostingEvent, errPostingEvent := http.DefaultClient.Do(req)

		if errPostingEvent != nil {
			log.Fatalf("ERROR: %s", errPostingEvent)
			fmt.Println("ERROR " + errPostingEvent.Error())
			return MyResponse{Status: false}, err
		}

		defer responsePostingEvent.Body.Close()
		body, _ := ioutil.ReadAll(responsePostingEvent.Body)

		fmt.Printf("Results: %s\n", string(body))
	*/
	return MyResponse{Status: true}, nil

}

func main() {
	lambda.Start(incrementUserEventBookings)
}

/*

API Gateway URL:
# request
{
	Bucket  string `json:"bucket"`
	Key     string `json:"user_uuid"` // The organizer of the event
	EventID string `json:"eventId"`  // Organizer of the event
}


# response
{
	"status" : bool
}
*/

/*
nlb_port 80
riak_cluster_nlb riak-cluster-network-lb-d90a8ac266b9ee92.elb.us-west-2.amazonaws.com


{
  "bucket": "eventbrite",
  "user_uuid": "4a6bd3f7-9fa9-44e7-a79c-2258594fe0c6",
  "eventId": "8ff53536-95df-42c5-8f31-8bf7733f2f7e"
}

*/
