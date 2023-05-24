package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Studentdetails struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Standard int    `json:"standard"`
	Division string `json:"division"`
}

type Subjectdetails struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type st_subject struct {
	Name       string `json:"name"`
	Standard   int    `json:"standard"`
	Division   string `json:"division"`
	SubjectIDs []int  `json:"subjectIds"`
}

type Student_subject struct {
	Id       int              `json:"id"`
	Name     string           `json:"name"`
	Standard int              `json:"standard"`
	Division string           `json:"division"`
	Subjects []Subjectdetails `json:"subjects"`
}

func main() {
	var err error
	db, err = sql.Open("postgres", "postgresql://max:roach@localhost:26257/school?sslmode=require")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", hello)
	r.Post("/students", poststudents)
	r.Put("/students", updatestudent)
	r.Get("/students", getstudents)
	r.Get("/studentsbydiv", getstudentsbydiv)
	r.Get("/studentsbysubjects", getstudentsbysubject)
	r.Get("/extra", extra)
	http.ListenAndServe(":3000", r)
}

func hello(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Hello World!"))
}

func poststudents(w http.ResponseWriter, r *http.Request) {
	var subject st_subject
	err := json.NewDecoder(r.Body).Decode(&subject)
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var id int

	query := `INSERT INTO student (name, standard, division)
              VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(query, subject.Name, subject.Standard, subject.Division).Scan(&id)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	for _, subjectID := range subject.SubjectIDs {
		query2 := `INSERT INTO student_subject (student_id, subject_id) VALUES ($1, $2)`
		_, err = tx.Exec(query2, id, subjectID)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte("Data inserted successfully"))
	fmt.Print("successfully inserted data in student details\n")
}

func updatestudent(w http.ResponseWriter, r *http.Request) {
	std := r.URL.Query().Get("std")
	id := r.URL.Query().Get("id")
	var student Studentdetails
	err := json.NewDecoder(r.Body).Decode((&student))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(
		"UPDATE student set standard = $1 Where id=$2 ", std, id); err != nil {
		log.Fatal(err)
	}
	w.Write([]byte("Data updated successfully"))
	fmt.Print("successfully updated data in student details\n")
}

func getstudents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from student")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	students := []Studentdetails{}
	for rows.Next() {
		var student Studentdetails
		if err := rows.Scan(&student.Id, &student.Name, &student.Standard, &student.Division); err != nil {
			log.Fatal(err)
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&students)
	fmt.Print("successfully got student details\n")
}

func getstudentsbydiv(w http.ResponseWriter, r *http.Request) {
	std := r.URL.Query().Get("std")
	div := r.URL.Query().Get("div")
	rows, err := db.Query("select * from student where standard=$1 and division=$2", std, div)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	students := []Studentdetails{}
	for rows.Next() {
		var student Studentdetails
		if err := rows.Scan(&student.Id, &student.Name, &student.Standard, &student.Division); err != nil {
			log.Fatal(err)
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(students)

	fmt.Print("successfully got student details filtered by division\n")
}

func getstudentsbysubject(w http.ResponseWriter, r *http.Request) {
	subject := r.URL.Query().Get("subject")
	rows, err := db.Query(`select student.id,student.name,student.standard ,student.division,subject.id ,subject.name from student left join student_subject on student_subject.student_id = student.id left join subject on subject.id = student_subject.subject_id where subject.name =$1 order by student.id`, subject)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	studentsubjects := make(map[int]Student_subject)
	for rows.Next() {
		var studentID int
		var studentName string
		var studentStandard int
		var studentDivision string
		var subjectID int
		var subjectName string

		if err := rows.Scan(&studentID, &studentName, &studentStandard, &studentDivision, &subjectID, &subjectName); err != nil {
			log.Fatal(err)
		}
		student_subject, ok := studentsubjects[studentID]
		if !ok {
			student_subject = Student_subject{
				Id:       studentID,
				Name:     studentName,
				Standard: studentStandard,
				Division: studentDivision,
			}
		}
		student_subject.Subjects = append(student_subject.Subjects, Subjectdetails{Id: subjectID, Name: subjectName})
		studentsubjects[studentID] = student_subject
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(studentsubjects)
	fmt.Print("successfully got student details filtered by subjects\n")
}

func extra(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from student")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	students := []Studentdetails{}
	for rows.Next() {
		var student Studentdetails
		if err := rows.Scan(&student.Id, &student.Name, &student.Standard, &student.Division); err != nil {
			log.Fatal(err)
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&students)
	fmt.Print("successfully got student details\n")
}

