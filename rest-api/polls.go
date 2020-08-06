package main

import (
	"errors"
	"net/http"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// poll defines the structure of a poll with 5 fields
type poll struct {
	ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
	APIKey  string         `json:"apikey"` // shouldn't be done in production
}

func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handlePollsGet(w, r)
		return
	case "POST":
		s.handlePollsPost(w, r)
		return
	case "DELETE":
		s.handlePollsDelete(w, r)
		return
	case "OPTIONS":
		// allow delete over CORS
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond(w, r, http.StatusOK, nil)
		return
	}
	// not found
	respondHTTPErr(w, r, http.StatusNotFound)
}

// Reading polls
func (s *Server) handlePollsGet(w http.ResponseWriter, r *http.Request) {
	var q *mgo.Query
	var result []*poll

	// Create copy of the database connection
	session := s.db.Copy()
	defer session.Close()

	// create an object referring to the polls collection
	c := session.DB("ballots").C("polls")

	// parse the url path into an instance of the Path type
	p := NewPath(r.URL.Path)

	// build an mgo.Query object by parsing the path
	if p.HasID() {
		q = c.FindId(bson.ObjectIdHex(p.ID)) // get a specific poll
	} else {
		q = c.Find(nil)	// get all polls
	}
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, errors.New("not implemented"))
		return
	}
	respond(w, r, http.StatusOK, &result)
}

// Creating a poll
func (s *Server) handlePollsPost(w http.ResponseWriter, r *http.Request) {
	var p poll

	// create a copy of the database connection
	session := s.db.Copy()
	defer session.Close()

	// create object referring to the polls collection
	c := session.DB("ballots").C("polls")

	// read the request body and store the value into &p
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "failed to read poll from request", err)
	}

	// Extract the apiKey
	apiKey, ok := APIKey(r.Context())
	if ok {
		p.APIKey = apiKey
	}
	p.ID = bson.NewObjectId()
	if err := c.Insert(p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "failed to insert poll", err)
		return
	}

	// point to the URL to access the newly created poll
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}

// Deleting a poll
func (s *Server) handlePollsDelete(w http.ResponseWriter, r *http.Request) {

	// create a copy of the database connection
	session := s.db.Copy()
	defer session.Close()

	// create an on=bject referring to the polls collection
	c := session.DB("ballots").C("polls")

	// parse the url path into an instance of the Path type
	p := NewPath(r.URL.Path)

	// check if the parsed path points to a poll
	// prevent deletion of all polls
	if !p.HasID() {
		respondErr(w, r, http.StatusMethodNotAllowed, "Cannot delete all polls!")
		return
	}

	// delete the poll with the given id and handle any errors
	if err := c.RemoveId(bson.ObjectIdHex(p.ID)); err != nil{
		respondErr(w, r, http.StatusInternalServerError, "failed to delete poll", err)
		return
	}
	respond(w, r, http.StatusOK, nil)	
}
