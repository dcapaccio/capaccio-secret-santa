package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"cloud.google.com/go/storage"
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

// "Global" vars -- should be in a daemon context or... well for now this is fine.
var (
	// dataFile = "https://storage.cloud.google.com/capaccio-secret-santa-2025/wishlist.json"
	mutex      sync.Mutex
	bucketName = "capaccio-secret-santa-2025"
	blobName   = "wishlist.json"
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

//  Example code, left in here for reference until stable.
// // MyCloudRunHandler Example is an http handler to call a update json function
// func MyCloudRunHandler(w http.ResponseWriter, r *http.Request) {

// 	// Example data to merge/update
// 	newUpdates := map[string]interface{}{
// 		"key_to_update":  "new_value_from_golang",
// 		"another_key_go": 456,
// 	}

// 	err := updateJSONInBucket(r.Context(), bucketName, blobName, newUpdates)
// 	if err != nil {
// 		log.Printf("Error updating JSON: %v", err)
// 		http.Error(w, fmt.Sprintf("Failed to update JSON: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	fmt.Fprintf(w, "JSON object %s in bucket %s updated successfully.\n", blobName, bucketName)
// }

// // updateJSONInBucket reads, updates, and writes a JSON file in a GCS bucket.
// func updateJSONInBucket(ctx context.Context, bucketName, blobName string, updates map[string]interface{}) error {
// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		return fmt.Errorf("storage.NewClient: %w", err)
// 	}
// 	defer client.Close()

// 	bucket := client.Bucket(bucketName)
// 	obj := bucket.Object(blobName)

// 	// 1. Read existing JSON (if applicable)
// 	// Create a reader to download the object
// 	rc, err := obj.NewReader(ctx)
// 	if err != nil {
// 		// If the file doesn't exist, we'll start with an empty map.
// 		// Handle other errors as actual failures.
// 		if err == storage.ErrObjectNotExist {
// 			log.Printf("Object %s does not exist, creating new.", blobName)
// 		} else {
// 			return fmt.Errorf("obj.NewReader: %w", err)
// 		}
// 	}
// 	defer rc.Close()

// 	var existingData map[string]interface{}
// 	if err == nil { // Only read if NewReader succeeded
// 		byteValue, err := ioutil.ReadAll(rc)
// 		if err != nil {
// 			return fmt.Errorf("ioutil.ReadAll: %w", err)
// 		}
// 		if len(byteValue) > 0 {
// 			if err := json.Unmarshal(byteValue, &existingData); err != nil {
// 				return fmt.Errorf("json.Unmarshal existing data: %w", err)
// 			}
// 		}
// 	}

// 	if existingData == nil {
// 		existingData = make(map[string]interface{})
// 	}

// 	// 2. Modify the JSON data
// 	for k, v := range updates {
// 		existingData[k] = v
// 	}

// 	// 3. Serialize the updated JSON
// 	updatedJSON, err := json.MarshalIndent(existingData, "", "  ") // Use MarshalIndent for pretty printing
// 	if err != nil {
// 		return fmt.Errorf("json.MarshalIndent: %w", err)
// 	}

// 	// 4. Upload the updated JSON
// 	wc := obj.NewWriter(ctx)
// 	wc.ContentType = "application/json"
// 	if _, err := wc.Write(updatedJSON); err != nil {
// 		return fmt.Errorf("wc.Write: %w", err)
// 	}
// 	if err := wc.Close(); err != nil {
// 		return fmt.Errorf("wc.Close: %w", err)
// 	}

// 	return nil
// }

// Load existing wishlists from file
func loadWishlists(ctx context.Context) ([]Wishlist, error) {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return []Wishlist{}, fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(blobName)

	// 1. Read existing JSON (if applicable)
	// Create a reader to download the object
	rc, err := obj.NewReader(ctx)
	if err != nil {
		// If the file doesn't exist, we'll start with an empty map.
		// Handle other errors as actual failures.
		if err == storage.ErrObjectNotExist {
			log.Printf("Object %s does not exist, creating new.", blobName)
		} else {
			return []Wishlist{}, fmt.Errorf("obj.NewReader: %w", err)
		}
	}
	defer rc.Close()

	var wishlists []Wishlist
	if err == nil { // Only read if NewReader succeeded
		err = json.NewDecoder(rc).Decode(&wishlists)
		if err != nil {
			return []Wishlist{}, fmt.Errorf("json.NewDecoder.Decode existing data: %w", err)
		}
	}
	return wishlists, nil
}

// Save wishlists to gcs storage as json file.
func saveWishlists(ctx context.Context, wishlists []Wishlist) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(blobName)

	// 3. Serialize the updated JSON
	updatedJSON, err := json.MarshalIndent(wishlists, "", "  ") // Use MarshalIndent for pretty printing
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}

	// 4. Upload the updated JSON
	wc := obj.NewWriter(ctx)
	wc.ContentType = "application/json"
	if _, err := wc.Write(updatedJSON); err != nil {
		return fmt.Errorf("wc.Write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("wc.Close: %w", err)
	}

	return nil
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

	wishlists, _ := loadWishlists(r.Context())

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

	saveWishlists(r.Context(), wishlists)

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

	wishlists, _ := loadWishlists(r.Context())

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
	wishlists, _ := loadWishlists(r.Context())

	tmpl := templates.Lookup("view.html.tmpl")
	tmpl.Execute(w, wishlists)
}

// get who you have page
func giftPageHandler(w http.ResponseWriter, r *http.Request) {
	wishlists, _ := loadWishlists(r.Context())

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

	wishlists, _ := loadWishlists(r.Context())
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
	http.HandleFunc("/gift", giftPageHandler)
	http.HandleFunc("/giftRecipient", giftRecipientHandler)
	// To serve CSS static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
