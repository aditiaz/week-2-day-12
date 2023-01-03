package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"my-web/connection"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var Data = map[string]interface{}{
	"Title":   "Personal Web",
	"IsLogin": true,
}

type dataInput struct{
	Id int
	ProjectName string
	Description string
	Technologies []string
	start_date time.Time
	end_date time.Time
	Duration string
	Image string

	


}

var dataInputs = []dataInput{
	{
	
	
		
	},
}


func main() {
	route := mux.NewRouter()
	connection.DataBaseConnection()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/",http.FileServer(http.Dir("./public"))))
	
	route.HandleFunc("/home",home).Methods("GET")
	route.HandleFunc("/home",addMyProject).Methods("POST")
	route.HandleFunc("/editProject/{id}",editProject).Methods("GET")
	route.HandleFunc("/updateProject/{id}",updateProject).Methods("POST")
	route.HandleFunc("/projectDetail/{id}",projectDetail).Methods("GET")
	route.HandleFunc("/contactMe",contactMe).Methods("GET")
	route.HandleFunc("/addProject",addProject).Methods("GET")
	route.HandleFunc("/delete-Project/{id}", deleteProject).Methods("GET")

	fmt.Println("Server is running on port 5000")
	http.ListenAndServe("localhost:5000", route)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	var result []dataInput
	rows, err := connection.Conn.Query(context.Background(), "SELECT id, name, description, technologies, start_date, end_date FROM tb_projects ")
	if err != nil {
		fmt.Println(err.Error())
		return 
	}

	for rows.Next() {
		var each = dataInput{}

		var err = rows.Scan(&each.Id, &each.ProjectName,&each.Description,&each.Technologies,&each.start_date, &each.end_date)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.Duration = period(each.start_date, each.end_date)
		result = append(result, each)
	}




	respData := map[string]interface{}{
		// "Data":  Data,
		"dataInputs": result,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,respData)
}

func projectDetail( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	id,_ := strconv.Atoi(mux.Vars(r)["id"])
	var tmpl,err = template.ParseFiles("views/detail-page.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}
	
	ProjectDetail := dataInput{}
	err =  connection.Conn.QueryRow(context.Background(),"SELECT id, name, description, technologies, start_date, end_date FROM tb_projects WHERE id=$1",id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName,&ProjectDetail.Description,&ProjectDetail.Technologies,&ProjectDetail.start_date, &ProjectDetail.end_date)
	    if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}
		ProjectDetail.Duration = period(ProjectDetail.start_date, ProjectDetail.end_date)
		Start_date :=   ProjectDetail.start_date.Format("02-Jan-2006")
		End_date :=   ProjectDetail.end_date.Format("02-Jan-2006")
	resp := map[string]interface{}{
		"dataInputs" : ProjectDetail,
		"start_date" : Start_date,
		"end_date" : End_date,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,resp)

}

func period(start time.Time, end time.Time)string{

	distance := end.Sub(start)

	var duration string
	year := int(distance.Hours()/(12 * 30 * 24))
	 if year != 0 {
		duration = strconv.Itoa(year) + " tahun"
	}else{
		month := int(distance.Hours()/(30 * 24))
		if month != 0 {
			duration = strconv.Itoa(month) + " bulan"
		}else{
			week := int(distance.Hours()/(7 *24))
			if week != 0 {
				duration = strconv.Itoa(week) +  " minggu"
			} else {
				day := int(distance.Hours()/(24))
				if day != 0 {
					duration = strconv.Itoa(day) + " hari"
				}
			}
		}
	}	
	return duration
}






func addMyProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}  
    projectName :=  r.PostForm.Get("name")
	description := r.PostForm.Get("description")
	checkbox := r.Form["checkbox"]
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)
	_,err = connection.Conn.Exec(context.Background(),"INSERT INTO tb_projects(name, description, technologies, start_date, end_date) VALUES ($1,$2,$3,$4,$5)",projectName,description,checkbox,startDateTime,endDateTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// fmt.Println(dataInputs)
	http.Redirect(w,r,"/home",http.StatusMovedPermanently)
}



func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}  
    projectName :=  r.PostForm.Get("name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("description")
	checkbox := r.Form["checkbox"]
	// image := r.PostForm.Get("image")
	Start_date,_ :=   time.Parse("2006-01-02",startDate)
	End_date,_ :=   time.Parse("2006-01-02",endDate)

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET name = $1, description = $2, technologies = $3, start_date = $4, end_date = $5 WHERE id=$6", projectName, description, checkbox, Start_date, End_date, id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
    }
	
		
	http.Redirect(w,r,"/home",http.StatusMovedPermanently)
}

func editProject( w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
    var tmpl,err = template.ParseFiles("views/update.html")
	id,_ := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}
	
	ProjectDetail := dataInput{}
	err =  connection.Conn.QueryRow(context.Background(),"SELECT id, name, description, technologies, start_date, end_date FROM tb_projects WHERE id=$1",id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName,&ProjectDetail.Description,&ProjectDetail.Technologies,&ProjectDetail.start_date, &ProjectDetail.end_date)
	    if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("message : " + err.Error()))
			return
		}
		Start_date :=   ProjectDetail.start_date.Format("01-02-2006")
		End_date :=   ProjectDetail.end_date.Format("01-02-2006")
	resp := map[string]interface{}{
		"dataInputs" : ProjectDetail,
		"start_date" : Start_date,
		"end_date" : End_date,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,resp)
}





func deleteProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	_,err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
	}

	http.Redirect(w, r, "/home", http.StatusFound)
}

func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/get-in-touch.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,Data)
}
func addProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/add-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w,Data)
}
