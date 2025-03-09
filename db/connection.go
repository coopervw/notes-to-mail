package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func GetDBConnection() *sql.DB {
	homedir, err := os.UserHomeDir()
	if (err != nil) {
		log.Fatal(err)
	}

	annotationDbSearchPatch := fmt.Sprintf("%s/Library/Containers/com.apple.iBooksX/Data/Documents/AEAnnotation", homedir)
	booksDbSearchPatch := fmt.Sprintf("%s/Library/Containers/com.apple.iBooksX/Data/Documents/BKLibrary", homedir)
	annotationsFname := findByExt(annotationDbSearchPatch)
	booksFname := findByExt(booksDbSearchPatch)

	var annotationDbPathWithoutPrefix string = fmt.Sprintf("%s/%s", annotationDbSearchPatch, annotationsFname)
	var bookDbPathWithoutPrefix string = fmt.Sprintf("%s/%s", booksDbSearchPatch, booksFname)

	var annotationDbPathWithPrefix string = fmt.Sprintf("file:%s/%s", annotationDbSearchPatch, annotationsFname)
	var bookDbPathWithPrefix string = fmt.Sprintf("file:%s/%s", booksDbSearchPatch, booksFname)

	if _, err := os.Stat(annotationDbPathWithoutPrefix); errors.Is(err, os.ErrNotExist) {
		log.Fatal("iBooks files are not found.")
	}
	if _, err := os.Stat(bookDbPathWithoutPrefix); errors.Is(err, os.ErrNotExist) {
		log.Fatal("iBooks files are not found.")
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s", bookDbPathWithPrefix))
	if err != nil {
		log.Fatal(err)
	}

	// Attach second SQLLite database file to connection
	_, err = db.Exec(fmt.Sprintf("attach database '%s' as a", annotationDbPathWithPrefix))
	if err != nil {
		log.Println(fmt.Sprintf("attach database '%s' as a", annotationDbPathWithPrefix))
		log.Fatal(err)
	}

	return db
}

func findByExt(path string) string {
	ext := ".sqlite$"
	var fname string
	filepath.Walk(path, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
				fname = f.Name()
			}
		}
		return nil
	})

	return fname
}

// FetchRandomNote gets a random note from all the books in the database
func FetchRandomNote(db *sql.DB) (*RandomNote, error) {
	// Step 1: Get all books and their note counts
	bookCountsRows, err := db.Query(GetBooksWithNoteCount)
	if err != nil {
		return nil, err
	}
	defer bookCountsRows.Close()

	// Step 2: Create a weighted list of book IDs
	type bookWeight struct {
		id         string
		noteCount  int
	}
	var books []bookWeight
	var totalNotes int

	for bookCountsRows.Next() {
		var bookID string
		var noteCount int
		err := bookCountsRows.Scan(&bookID, &noteCount)
		if err != nil {
			return nil, err
		}
		books = append(books, bookWeight{id: bookID, noteCount: noteCount})
		totalNotes += noteCount
	}

	if err = bookCountsRows.Err(); err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return nil, sql.ErrNoRows
	}

	// Step 3: Select a random book, weighted by note count
	rand.Seed(time.Now().UnixNano())
	randomValue := rand.Intn(totalNotes) + 1
	var selectedBookID string

	runningTotal := 0
	for _, book := range books {
		runningTotal += book.noteCount
		if randomValue <= runningTotal {
			selectedBookID = book.id
			break
		}
	}

	// Step 4: Get book information
	var bookTitle, bookAuthor string
	row := db.QueryRow(GetBookDataById, selectedBookID)
	err = row.Scan(&bookTitle, &bookAuthor)
	if err != nil {
		return nil, err
	}

	// Step 5: Get a random note from the selected book
	noteRow := db.QueryRow(GetRandomNoteFromBook, selectedBookID)
	
	var highlight string
	var note sql.NullString
	err = noteRow.Scan(&highlight, &note)
	if err != nil {
		return nil, err
	}

	// Step 6: Create the RandomNote object
	randomNote := &RandomNote{
		BookId:     selectedBookID,
		BookTitle:  bookTitle,
		BookAuthor: bookAuthor,
		Highlight:  highlight,
	}
	
	if note.Valid {
		randomNote.Note = note.String
	}

	return randomNote, nil
}
