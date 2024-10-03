package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"
	
    "github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Ticket type and Event struct definition
type Ticket struct {
    Type      string  `json:"type"`
    Price     float64 `json:"price"`
    Available int     `json:"available"`
}

type Event struct {
    Title   string   `json:"title"`
    Date    string   `json:"date"`
    Tickets []Ticket `json:"tickets"`
}

// Order struct to store ticket purchases
type Order struct {
    ID              int       `json:"id"`
    Username        string    `json:"username"`
    GeneralAdmission int      `json:"general_admission"`
    VIP             int       `json:"vip"`
    TotalCost       float64   `json:"total_cost"`
    PurchaseTime    time.Time `json:"purchase_time"`
}

var (
    event = Event{
        Title: "Rock Concert",
        Date:  "2024-12-31",
        Tickets: []Ticket{
            {Type: "general_admission", Price: 50.0, Available: 100},
            {Type: "vip", Price: 150.0, Available: 50},
        },
    }
    orders          []Order
    mu              sync.Mutex
    orderIDCounter  = 1
)

// Handler to show available tickets for an event
func eventDetailsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(event)
}

// Handler to review tickets before finalizing purchase
func reviewTicketHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var purchaseRequest map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&purchaseRequest); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    generalAdmission := int(purchaseRequest["general_admission"].(float64))
    vip := int(purchaseRequest["vip"].(float64))

    if generalAdmission > event.Tickets[0].Available || vip > event.Tickets[1].Available {
        http.Error(w, "Not enough tickets available", http.StatusBadRequest)
        return
    }

    totalCost := float64(generalAdmission)*event.Tickets[0].Price + float64(vip)*event.Tickets[1].Price

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":    "Review your order",
        "total_cost": totalCost,
    })
}

// Handler to finalize ticket booking
func bookTicketHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var purchaseRequest map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&purchaseRequest); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    generalAdmission := int(purchaseRequest["general_admission"].(float64))
    vip := int(purchaseRequest["vip"].(float64))

    if generalAdmission > event.Tickets[0].Available || vip > event.Tickets[1].Available {
        http.Error(w, "Not enough tickets available", http.StatusBadRequest)
        return
    }

    totalCost := float64(generalAdmission)*event.Tickets[0].Price + float64(vip)*event.Tickets[1].Price

    // Update availability
    event.Tickets[0].Available -= generalAdmission
    event.Tickets[1].Available -= vip

    // Create a new order entry
    order := Order{
        ID:              orderIDCounter,
        Username:        "anonymous", // Replace with actual username once authentication is implemented
        GeneralAdmission: generalAdmission,
        VIP:             vip,
        TotalCost:       totalCost,
        PurchaseTime:    time.Now(),
    }
    orders = append(orders, order)
    orderIDCounter++

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":    "Tickets booked successfully",
        "total_cost": totalCost,
    })
}

// Handler to cancel a ticket
func cancelTicketHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var cancelRequest map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&cancelRequest); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    ticketType := cancelRequest["ticket_type"].(string)
    quantity := int(cancelRequest["quantity"].(float64))

    mu.Lock()
    defer mu.Unlock()

    // Refund logic based on ticket type
    if ticketType == "general_admission" {
        event.Tickets[0].Available += quantity
    } else if ticketType == "vip" {
        event.Tickets[1].Available += quantity
    } else {
        http.Error(w, "Invalid ticket type", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Ticket cancellation successful",
    })
}

// Handler to retrieve order history
func orderHistoryHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}


func main() {
    router := mux.NewRouter()
    
    // Define the route for fetching place details
    router.HandleFunc("/event-details", eventDetailsHandler)
    router.HandleFunc("/review-ticket", reviewTicketHandler)
    router.HandleFunc("/confirm-ticket", bookTicketHandler)
    router.HandleFunc("/cancel-ticket", cancelTicketHandler)
    router.HandleFunc("/order-history", orderHistoryHandler)
	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Update with your frontend origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})

	// Start the server on port 4000
	log.Println("Server running on port 4000")
	log.Fatal(http.ListenAndServe("localhost:4000", c.Handler(router)))
}