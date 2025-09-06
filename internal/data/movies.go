package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`                      //unique integer ID
	CreatedAt time.Time `json:"-"`                       // Timestamp for when the movie is added to our database // this isn't relevant for end users
	Title     string    `json:"title"`                   //Movie title
	Year      int32     `json:"year,omitzero"`           // Movie release year
	Runtime   Runtime   `json:"runtime,omitzero,string"` //movie runtime in minutes
	Genres    []string  `json:"genres,omitempty"`        //Slice of genres for the movie
	Version   int32     `json:"version"`                 // starts at 1 and will be incremented each time the movie information is updated
}
