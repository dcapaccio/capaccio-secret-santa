package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
)

type Item struct {
	Name     string `json:"name"`
	Priority string `json:"priority"`
	Link     string `json:"link,omitempty"`
}

type Wishlist struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Items     []Item `json:"items"`
	BuyingFor string `json:"buyingFor,omitempty"`
	Household string `json:"household,omitempty"`
}

var (
	dataFile = "gs://capaccio-secret-santa-2025/wishlist.json"
	mutex    sync.Mutex
)

// Predefined people with unique IDs
var people = map[string]int{
	"Ben":      1,
	"Teresa":   2,
	"Jen":      3,
	"Chris":    4,
	"Danny":    5,
	"Marianna": 6,
	"Alyse":    7,
	"John":     8,
	"Laura":    9,
	"Vincent":  10,
	"Joe":      11,
}

// Load existing wishlists from file
func loadWishlists() ([]Wishlist, error) {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Wishlist{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var wishlists []Wishlist
	err = json.NewDecoder(file).Decode(&wishlists)
	if err != nil {
		return []Wishlist{}, nil
	}
	return wishlists, nil
}

// Save wishlists to file
func saveWishlists(wishlists []Wishlist) error {
	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(wishlists)
}

// Handle form submission
func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	id, ok := people[name]
	if !ok {
		http.Error(w, "Invalid name selected", http.StatusBadRequest)
		return
	}

	// Collect items
	var items []Item
	itemNames := r.Form["itemName"]
	priorities := r.Form["priority"]
	links := r.Form["link"]

	for i := range itemNames {
		if itemNames[i] != "" {
			// if the items list already has an item with the same name, skip it
			duplicate := false
			for _, existingItem := range items {
				if existingItem.Name == itemNames[i] {
					duplicate = true
					break
				}
			}
			if !duplicate {
				items = append(items, Item{
					Name:     itemNames[i],
					Priority: priorities[i],
					Link:     links[i],
				})
			}
		}
	}

	mutex.Lock()
	defer mutex.Unlock()

	wishlists, _ := loadWishlists()

	// Check if person's name exists
	found := false
	for i := range wishlists {
		if wishlists[i].Name == name {
			wishlists[i].Items = append(wishlists[i].Items, items...)
			found = true
			break
		}
	}

	if !found {
		wishlists = append(wishlists, Wishlist{ID: id, Name: name, Items: items})
	}

	saveWishlists(wishlists)

	http.Redirect(w, r, "/view", http.StatusSeeOther)
}

// Load templates from separate .html.tmpl files
var templates = template.Must(template.ParseFiles("form.html.tmpl", "view.html.tmpl", "viewRecipient.html.tmpl", "giftPage.html.tmpl"))

// Render form
func formHandler(w http.ResponseWriter, r *http.Request) {
	// Build dropdown options dynamically
	var names []string
	for name := range people {
		names = append(names, name)
	}

	wishlists, _ := loadWishlists()

	tmplData := struct {
		NameOptions []string
		Wishlist    []Wishlist
	}{
		NameOptions: names,
		Wishlist:    wishlists,
	}

	tmpl := templates.Lookup("form.html.tmpl")
	tmpl.Execute(w, tmplData)
}

// Render wishlists
func viewHandler(w http.ResponseWriter, r *http.Request) {
	wishlists, _ := loadWishlists()

	tmpl := templates.Lookup("view.html.tmpl")
	tmpl.Execute(w, wishlists)
}

// API endpoint: return JSON
func apiHandler(w http.ResponseWriter, r *http.Request) {
	wishlists, _ := loadWishlists()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wishlists)
}

// get who you have page
func giftPageHandler(w http.ResponseWriter, r *http.Request) {
	wishlists, _ := loadWishlists()

	// Build dropdown options dynamically
	var names []string
	for _, wl := range wishlists {
		names = append(names, wl.Name)
	}

	tmplData := struct {
		NameOptions []string
	}{
		NameOptions: names,
	}

	tmpl := templates.Lookup("giftPage.html.tmpl")
	tmpl.Execute(w, tmplData)
}

// Gift recipient lookup Page
func giftRecipientHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}

	wishlists, _ := loadWishlists()
	var recipientName string
	var recipientWishlist []Item

	// Find who this person is buying for
	for _, wl := range wishlists {
		if wl.Name == name {
			recipientName = wl.BuyingFor
			break
		}
	}
	// List the recipient's wishlist
	for _, wl := range wishlists {
		if wl.Name == recipientName {
			recipientWishlist = wl.Items
			break
		}
	}

	if recipientName == "" {
		w.Write([]byte("No gift recipient found."))
		return
	}

	tmplData := struct {
		Recipient string
		Items     []Item
	}{
		Recipient: recipientName,
		Items:     recipientWishlist,
	}

	tmpl := templates.Lookup("viewRecipient.html.tmpl")
	tmpl.Execute(w, tmplData)
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/api/wishlists", apiHandler)
	http.HandleFunc("/gift", giftPageHandler)
	http.HandleFunc("/giftRecipient", giftRecipientHandler)
	// To serve CSS static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
