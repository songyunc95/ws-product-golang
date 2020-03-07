package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	// "sort"
	"strings"
	"os"
)

type values struct {
	view  int 
	click int 
}

type counters struct {
	sync.Mutex
	m map[string]*values
}

type rateLimit struct {
	sync.Mutex
	prev 	 time.Time
//	interval float64
//	limit 	 int
	count    int
}

var (
	c = counters{m: make(map[string]*values)}
	rl = rateLimit{prev: time.Now(), count: 0}

	content = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// Using RFC822 as the format for timestamp
	data := fmt.Sprintf(content[rand.Intn(len(content))] + ":" + time.Now().Format(time.RFC822))
	
	c.Lock()
	// increment view
	if _, exists := c.m[data]; exists {
		(*c.m[data]).view += 1
	// init c.m[data] if c.m[data] does not exist
	} else {
		c.m[data] = &values{1, 0}
	}
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}
	
	/*
	// For testing
	var keys []string
	for k := range c.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf( "%v\n\tview: %v\tclick: %v\n", k, (*c.m[k]).view, (*c.m[k]).click)
	}
	*/
	
	
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	c.Lock()
	// increment click
	(*c.m[data]).click += 1
	c.Unlock()

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed(10, 5) {
		w.WriteHeader(429)
		return
	}
}

func isAllowed(limit int, interval float64) bool {
	rl.Lock()
	defer rl.Unlock()
	
	elapsed := time.Since(rl.prev).Seconds()
	
	// Checking if a unit of time interval is passed
	if elapsed >= interval {
		rl.count = 0
		rl.prev = time.Now()
	}
	
	// increment call_count because inputted function is called.
	rl.count++
		
	// Checking if the number of function calls exceeds the call limit
	if rl.count > limit {
		// For Testing
		// fmt.Printf("False, count: %v, limit: %v, interval: %v\n", rl.count, limit, interval)
		return false
	}
	// For Testing
	// fmt.Printf("True, count: %v, limit: %v, interval: %v\n", rl.count, limit, interval)
	return true
}

func uploadCounters() error {
	// Saving counters to file (as a mock store)
	// init file
	f, err := os.Create("store.json")
    if err != nil {
        return err
    }
	defer f.Close()
	
	var json_list []string
	// iterate over c.m to create json string for each key-value pair
	for k, v := range c.m {
		json_list = append(json_list, fmt.Sprintf("\"%s\":{\"view\": %v, \"click\": %v}", k, (*v).view, (*v).click))
	}
	// Join key-value pair json string
	json_str := "{" + strings.Join(json_list[:],",") + "}"
	
	// Save to file
	_, err = f.WriteString(json_str)
    if err != nil {
        return err
    }	
	
	return nil
}

func uploadEvery(sec int) {
	for {
		// Wait for sec Seconds
		time.Sleep(time.Duration(sec) * time.Second)
		go uploadCounters()
	}
}

func main() {
	// save mock store every 5 secs
	go uploadEvery(5)
	
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
