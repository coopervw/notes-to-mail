package db

import (
	"database/sql"
)

type SingleBook struct {
	Name   string
	Author string
}

type SingleHighlightNote struct {
	HightLight string
	Note       sql.NullString
}

type SingleBookInList struct {
	Id     string
	Title  string
	Author string
	Number int
}

// RandomNote represents a randomly selected note with its book information
type RandomNote struct {
	BookId      string
	BookTitle   string
	BookAuthor  string
	Highlight   string
	Note        string
}
