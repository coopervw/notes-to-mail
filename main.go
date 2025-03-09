package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
	dbThings "ibooks_notes_exporter/db"
	"ibooks_notes_exporter/email"
	"log"
	"os"
	"strings"
	"unicode"
)

func main() {
	app := &cli.App{
		Name:    "Ibooks notes exporter",
		Usage:   "Export your records from Apple iBooks",
		Authors: []*cli.Author{{Name: "Andrey Korchak", Email: "me@akorchak.software"}},
		Version: "v0.0.5",
		Commands: []*cli.Command{
			{
				Name:   "books",
				Usage:  "Get list of the books with notes and highlights",
				Action: getListOfBooks,
			},
			{
				Name: "version",
				Action: func(context *cli.Context) error {
					fmt.Printf("%s\n", context.App.Version)
					return nil
				},
			},
			{
				Name:      "export",
				HideHelp:  false,
				Usage:     "Export all notes and highlights from book with [BOOK_ID]",
				UsageText: "Export all notes and highlights from book with [BOOK_ID]",
				Action:    exportNotesAndHighlights,
				ArgsUsage: "ibooks_notes_exporter export BOOK_ID_GOES_HERE",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "book_id",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "skip_first_x_notes",
						Value:    0,
						Required: false,
					},
				},
				},
				{
					Name:      "mail",
					Usage:     "Send a random note to the configured email address",
					Action:    mailRandomNote,
					ArgsUsage: "ibooks_notes_exporter mail",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "to",
							Usage:   "Email address to send to (overrides environment variable)",
							Aliases: []string{"t"},
						},
						&cli.StringFlag{
							Name:    "from",
							Usage:   "Email address to send from (overrides environment variable)",
							Aliases: []string{"f"},
						},
						&cli.StringFlag{
							Name:    "api-key",
							Usage:   "Sendgrid API key (overrides environment variable)",
							Aliases: []string{"k"},
						},
					},
				},
			},
		}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

// mailRandomNote is the handler for the mail command
func mailRandomNote(cCtx *cli.Context) error {
	// Get DB connection
	db := dbThings.GetDBConnection()
	defer db.Close()

	// Get a random note
	randomNote, err := dbThings.FetchRandomNote(db)
	if err != nil {
		return fmt.Errorf("failed to get a random note: %v", err)
	}

	// Display the note that will be sent
	fmt.Println("Selected random note:")
	fmt.Printf("Book: %s — %s\n", randomNote.BookTitle, randomNote.BookAuthor)
	fmt.Printf("> %s\n", strings.Replace(randomNote.Highlight, "\n", "", -1))
	if randomNote.Note != "" {
		fmt.Printf("\n%s\n", strings.Replace(randomNote.Note, "\n", "", -1))
	}
	
	// Get email config from environment variables
	config := email.GetConfigFromEnv()
	
	// Override with command line args if provided
	if cCtx.String("to") != "" {
		config.ToEmail = cCtx.String("to")
	}
	if cCtx.String("from") != "" {
		config.FromEmail = cCtx.String("from")
	}
	if cCtx.String("api-key") != "" {
		config.APIKey = cCtx.String("api-key")
	}
	
	// Check if recipient email is set
	if config.ToEmail == "" {
		return fmt.Errorf("recipient email is required. Set EMAIL_TO_ADDRESS environment variable or use --to flag")
	}
	
	// Send the email
	fmt.Printf("\nSending email to %s...\n", config.ToEmail)
	err = email.SendRandomNote(config, *randomNote)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	
	fmt.Println("Email sent successfully!")
	return nil
}

func GetLastName(name string) string {
	// Split the input string into words
	words := strings.Fields(name)

	// Search backwards from the end of the string for the first non-title word
	var lastName string
	for i := len(words) - 1; i >= 0; i-- {
		if !isHonorific(words[i]) {
			lastName = words[i]
			break
		}
	}

	// Remove any trailing commas or periods from the last name
	lastName = strings.TrimSuffix(lastName, ",")
	lastName = strings.TrimSuffix(lastName, ".")

	// Return the last name in parentheses
	return "(" + lastName + ")"
}

// Helper function to check if a word is an honorific title
func isHonorific(word string) bool {
	return len(word) <= 3 && unicode.IsUpper(rune(word[0])) && (word[len(word)-1] == '.' || word[len(word)-1] == ',')
}

func GetLastNames(names string) string {
	// Split the input string into individual names
	nameList := strings.Split(names, " & ")

	// If there is only one name, just return the last name
	if len(nameList) == 1 {
		return GetLastName(nameList[0])
	}

	// If there are two names, combine the last names with "&"
	if len(nameList) == 2 {
		return GetLastName(nameList[0]) + " & " + GetLastName(nameList[1])
	}

	// If there are more than two names, combine the first name and last names with "&"
	firstName := nameList[0]
	lastNames := make([]string, len(nameList)-1)
	for i, name := range nameList[1:] {
		lastNames[i] = GetLastName(name)
	}
	return GetLastName(firstName) + " & " + strings.Join(lastNames, " & ")
}

func getListOfBooks(cCtx *cli.Context) error {
	db := dbThings.GetDBConnection()

	// Getting a list of books
	rows, err := db.Query(dbThings.GetAllBooksDbQueryConstant)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Render table with books
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"SingleBook ID", "# notes", "Title and Author"})

	var singleBook dbThings.SingleBookInList
	for rows.Next() {
		err := rows.Scan(&singleBook.Id, &singleBook.Title, &singleBook.Author, &singleBook.Number)
		if err != nil {
			log.Fatal(err)
		}
		// truncate title as needed so that table doesn't wrap when terminal width is narrow
		truncatedTitle := singleBook.Title
		if len(singleBook.Title) > 30 {
			truncatedTitle = singleBook.Title[:30] + "..."
		}
		// shortened author name(s)
		standardizedAuthor := GetLastNames(singleBook.Author)
		// The title and author looks like: "My Great Book (Doe)"
		t.AppendRows([]table.Row{
			{singleBook.Id, singleBook.Number, fmt.Sprintf("%s %s", truncatedTitle, standardizedAuthor)},
		})
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	t.Render()
	return nil
}

func exportNotesAndHighlights(cCtx *cli.Context) error {
	db := dbThings.GetDBConnection()
	defer db.Close()

	bookId := cCtx.String("book_id")
	skipXNotes := cCtx.Int("skip_first_x_notes")
	fmt.Println(bookId)

	var book dbThings.SingleBook
	row := db.QueryRow(dbThings.GetBookDataById, bookId)
	err := row.Scan(&book.Name, &book.Author)
	if err != nil {
		//log.Fatal()
		log.Println(err)
		log.Fatal("SingleBook is not found in iBooks!")
	}

	// Render MarkDown into STDOUT
	fmt.Println(fmt.Sprintf("# %s — %s\n", book.Name, book.Author))

	rows, err := db.Query(dbThings.GetNotesHighlightsById, bookId, skipXNotes)
	if err != nil {
		log.Fatal(err)
	}

	var singleHightLightNote dbThings.SingleHighlightNote
	for rows.Next() {
		err := rows.Scan(&singleHightLightNote.HightLight, &singleHightLightNote.Note)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(fmt.Sprintf("> %s", strings.Replace(singleHightLightNote.HightLight, "\n", "", -1)))

		if singleHightLightNote.Note.Valid {
			fmt.Println(fmt.Sprintf("\n%s", strings.Replace(singleHightLightNote.Note.String, "\n", "", -1)))
		}

		fmt.Println("---\n\n")

	}

	return nil
}
