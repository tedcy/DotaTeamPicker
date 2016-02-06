package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

//apt-get install mysql-server
//create database DOTAMATCH;
//use DOTAMATCH;
/*CREATE TABLE `MatchWithPlayers` ( 
	`matchId` int(11) NOT NULL, 
	`playerIds` int(11) NOT NULL, 
	PRIMARY KEY (`matchId`) 
);*/
//delete from MatchWithPlayers where matchId = 101;
func main(){
	db, err := sql.Open("mysql","root:password@tcp(127.0.0.1:3306)/DOTAMATCH")
	fmt.Printf("%T\n",db)

	if err != nil {
		fmt.Printf("Open database error: %s\n", err)
    }
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }

	stmtIns, err := db.Prepare("INSERT INTO MatchWithPlayers VALUES( ?, ? )")

	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }
	defer stmtIns.Close()

	stmtOut, err := db.Prepare("SELECT * FROM MatchWithPlayers WHERE matchId = ?")

	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }
	defer stmtOut.Close()

	_ ,err = stmtIns.Exec(101,10086)

	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }

	rows,err := stmtOut.Query(101)

	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }

	defer rows.Close()

	var matchId int
	var playerIds int
	for rows.Next() {
		err := rows.Scan(&matchId, &playerIds)

		if err != nil {
			fmt.Printf("%s\n", err)
        }
		fmt.Println(matchId,playerIds)
    }

	err = rows.Err()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
    }
}
