package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Simple in-memory inventory for donuts.
type Donut struct {
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Inventory struct {
	mu    sync.RWMutex
	items map[string]*Donut
}

func NewInventory() *Inventory {
	return &Inventory{
		items: map[string]*Donut{
			"glazed":    {Type: "glazed", Price: 1.25, Quantity: 100},
			"chocolate": {Type: "chocolate", Price: 1.50, Quantity: 80},
			"jelly":     {Type: "jelly", Price: 1.75, Quantity: 50},
		},
	}
}

func (inv *Inventory) listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	inv.mu.RLock()
	defer inv.mu.RUnlock()

	out := make([]*Donut, 0, len(inv.items))
	for _, d := range inv.items {
		out = append(out, &Donut{Type: d.Type, Price: d.Price, Quantity: d.Quantity})
	}
	writeJSON(w, out)
}

type tradeRequest struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type tradeResponse struct {
	Type      string  `json:"type"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	Remaining int     `json:"remaining"`
	Message   string  `json:"message,omitempty"`
}

func (inv *Inventory) buyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req tradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 || req.Type == "" {
		http.Error(w, "type and positive quantity required", http.StatusBadRequest)
		return
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	d, ok := inv.items[req.Type]
	if !ok {
		http.Error(w, "donut type not found", http.StatusNotFound)
		return
	}
	if req.Quantity > d.Quantity {
		http.Error(w, "not enough stock", http.StatusBadRequest)
		return
	}
	d.Quantity -= req.Quantity
	total := float64(req.Quantity) * d.Price

	writeJSON(w, tradeResponse{
		Type:      d.Type,
		Quantity:  req.Quantity,
		Total:     total,
		Remaining: d.Quantity,
		Message:   "purchase successful",
	})
}

func (inv *Inventory) sellHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req tradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 || req.Type == "" {
		http.Error(w, "type and positive quantity required", http.StatusBadRequest)
		return
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	// If type doesn't exist, create it with a default price.
	d, ok := inv.items[req.Type]
	if !ok {
		inv.items[req.Type] = &Donut{Type: req.Type, Price: 1.00, Quantity: req.Quantity}
		writeJSON(w, tradeResponse{
			Type:      req.Type,
			Quantity:  req.Quantity,
			Total:     0,
			Remaining: req.Quantity,
			Message:   "new donut type created and sold to inventory (price defaulted to 1.00)",
		})
		return
	}

	d.Quantity += req.Quantity
	writeJSON(w, tradeResponse{
		Type:      d.Type,
		Quantity:  req.Quantity,
		Total:     0,
		Remaining: d.Quantity,
		Message:   "sell to inventory successful",
	})
}

type createRequest struct {
	Type  string  `json:"type"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func (inv *Inventory) createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Type == "" || req.Price <= 0 || req.Stock < 0 {
		http.Error(w, "invalid fields", http.StatusBadRequest)
		return
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	if _, exists := inv.items[req.Type]; exists {
		http.Error(w, "donut type already exists", http.StatusBadRequest)
		return
	}
	inv.items[req.Type] = &Donut{Type: req.Type, Price: req.Price, Quantity: req.Stock}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, inv.items[req.Type])
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func main() {
	inv := NewInventory()

	http.HandleFunc("/inventory", inv.listHandler)   // GET
	http.HandleFunc("/donuts/buy", inv.buyHandler)   // POST {type, quantity}
	http.HandleFunc("/donuts/sell", inv.sellHandler) // POST {type, quantity}
	http.HandleFunc("/donuts", inv.createHandler)    // POST {type, price, stock}

	log.Println("donut API listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
