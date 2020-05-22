// Copyright 2020 John McKenzie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	listenAddress = "0.0.0.0:4778"
	version       = "0.0.1"
)

// Widget represents a generic object.
type Widget struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Description string `json:"description"`
}

// WidgetHandler handles Widget requests.
type WidgetHandler struct {
	widgets map[string]Widget
}

// NewWidgetHandler will construct a new WidgetHandler.
func NewWidgetHandler() WidgetHandler {
	return WidgetHandler{
		widgets: make(map[string]Widget, 0),
	}
}

func (h WidgetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.EscapedPath()
	id := strings.Replace(path, "/widgets/", "", 1)
	log.Printf("path: %s method: %s id: %s", path, r.Method, id)

	switch r.Method {
	case http.MethodGet:
		if len(id) > 0 {
			h.get(w, r, id)
		} else {
			h.list(w, r)
		}
		return
	case http.MethodPost:
		if len(id) > 0 {
			break
		}
		h.create(w, r)
		return
	case http.MethodPut:
		if len(id) <= 0 {
			break
		}
		h.update(w, r, id)
		return
	case http.MethodDelete:
		if len(id) <= 0 {
			break
		}
		h.delete(w, r, id)
		return
	default:
		// default method not allowed...
	}

	writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed for this resource.")
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/widgets/", NewWidgetHandler())

	log.Printf("listening for connections at %s", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Printf("URL Path: %s Method: %s", r.URL.Path, r.Method)
	if r.URL.Path != "/" {
		writeJSONError(w, http.StatusNotFound, "The requested resource could not be located.")
		return
	}

	if r.Method != http.MethodOptions && r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed for this resource.")
		return
	}

	payload := map[string]string{
		"timestamp": time.Now().String(),
		"version":   version,
	}

	if err := writeJSON(w, http.StatusOK, payload); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h WidgetHandler) list(w http.ResponseWriter, r *http.Request) {
	widgets := make([]Widget, 0)
	for _, widget := range h.widgets {
		widgets = append(widgets, widget)
	}

	payload := map[string]interface{}{
		"widgets": widgets,
		"count":   len(widgets),
	}

	if err := writeJSON(w, http.StatusOK, payload); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h WidgetHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	widget, ok := h.widgets[id]
	if !ok {
		log.Printf("unable to find widget with id %s", id)
		writeJSONError(w, http.StatusNotFound, "The requested resource could not be located.")
		return
	}

	if err := writeJSON(w, http.StatusOK, map[string]Widget{"widget": widget}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h WidgetHandler) create(w http.ResponseWriter, r *http.Request) {
	var widget Widget

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&widget); err != nil {
		log.Printf("unable to parse widget %s", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Printf("unable to generate uuid %x", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	widget.ID = strings.TrimSpace(string(uuid))
	h.widgets[widget.ID] = widget

	if err := writeJSON(w, http.StatusCreated, map[string]Widget{"widget": widget}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h WidgetHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	widget, ok := h.widgets[id]
	if !ok {
		log.Printf("unable to find widget with id %s", id)
		writeJSONError(w, http.StatusNotFound, "The requested resource could not be located.")
		return
	}

	var updWidget Widget
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updWidget); err != nil {
		log.Printf("unable to parse widget %s", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	widget.Name = updWidget.Name
	widget.Description = updWidget.Description
	h.widgets[widget.ID] = widget

	if err := writeJSON(w, http.StatusOK, map[string]Widget{"widget": widget}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func (h WidgetHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	widget, ok := h.widgets[id]
	if !ok {
		log.Printf("unable to find widget with id %s", id)
		writeJSONError(w, http.StatusNotFound, "The requested resource could not be located.")
		return
	}
	delete(h.widgets, id)

	if err := writeJSON(w, http.StatusOK, map[string]Widget{"widget": widget}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) error {
	log.Printf("writing json response code %d with payload %s", status, payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, map[string]string{
		"error": message,
	})
}
