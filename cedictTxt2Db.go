package main

import (
  "database/sql"
  "fmt"
  _ "github.com/mattn/go-sqlite3"
  "os"
  "io/ioutil"
  "strings"
  "regexp"
)

// regexp used to parse the lines in the raw, downloaded text file
var parseRegexp, _ = regexp.Compile(`(.*?) (.*?) \[(.*?)\] \/(.*)\/`)

func main() {
  if len(os.Args) != 2 {
    fmt.Println("Usage: cedictTxt2Db <input text file>")
    os.Exit(-1)
  }

  fmt.Println("Load raw data...")

  // load raw data into memory
  txtData, err := ioutil.ReadFile(os.Args[1])
  if err != nil {
    fmt.Println(err)
    return
  }

  // split data into lines
  lines := strings.Split(string(txtData), "\n")


  // remove old database
  os.Remove("./cedict.sqlite3")

  fmt.Println("Create database...")

  // create new database file and defer closing it
  db, err := sql.Open("sqlite3", "./cedict.sqlite3")
  if err != nil {
    fmt.Println(err)
    return
  }
  defer db.Close()

  // SQL code to create the table
  createTableSql := []string {
    `CREATE TABLE dict (      
      trad    VARCHAR(50),      
      simp    VARCHAR(50),     
      pinyin  VARCHAR(100),     
      english VARCHAR(500))`,
    "DELETE FROM dict",
  }

  // create table
  for _, sql := range createTableSql {
    _, err = db.Exec(sql)
    if err != nil {
      fmt.Printf("%q: %s\n", err, sql)
      return
    }
  }

  // create new transaction to add the data to the databse
  tx, err := db.Begin()
  if err != nil {
    fmt.Println(err)
    return
  }

  // prepare insert statement and defer closing it
  stmt, err := tx.Prepare("INSERT INTO dict(trad, simp, pinyin, english) VALUES(?, ?, ?, ?)")
  if err != nil {
    fmt.Println(err)
    return
  }
  defer stmt.Close()

  fmt.Println("Add data to database...")

  // loop through all the lines in the text file
  for _,line := range lines {
    // only parse non-empty lines and those that are not comments (start with '#')
    if (line != "" && !(strings.HasPrefix(line, "#"))) {
      // use compiled regexp to parse line
      parseResult := parseRegexp.FindStringSubmatch(line)

      // execute insert statement for this row
      _, err = stmt.Exec(parseResult[1], // trad
        parseResult[2], // simp
        strings.Replace(parseResult[3], "u:", "v", -1), // pinyin (replace "u:" by "v")
        strings.Replace(parseResult[4], "/", " / ", -1)) // english (replace "/" by " / ")
      if err != nil {
        fmt.Println(err)
        return
      }
    }
  }

  // commit transation
  tx.Commit()

  fmt.Println("Create indices...")

  // SQL code to add indices to trad and simp columns
  indexSql := []string{
    "CREATE INDEX trad_index ON dict(trad)",
    "CREATE INDEX simp_index ON dict(simp)",
  }

  // execute the above sql code
  for _, sql := range indexSql {
    _, err = db.Exec(sql)
    if err != nil {
      fmt.Printf("%q: %s\n", err, sql)
      return
    }
  }
}



