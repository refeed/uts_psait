package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DBPATH = "refeeds:12345678@/sait_db_uts"
)

func main() {
	http.HandleFunc("/nilaiMahasiswa", nilaiMahasiswaHandler)
	http.ListenAndServe(":8080", nil)
}

func nilaiMahasiswaHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		nim := r.URL.Query().Get("nim")
		if nim != "" {
			fmt.Fprint(w, getNilaiMahasiswaJSON(nim))
			return
		}

		fmt.Fprint(w, getNilaiMahasiswasJSON())
		return
	case "POST":
		handleNilaiMahasiswaPost(w, r)
	case "PATCH":
		handleNilaiMahasiswaPatch(w, r)
	case "DELETE":
		handleNilaiMahasiswaDelete(w, r)
	}
}

func handleNilaiMahasiswaDelete(w http.ResponseWriter, r *http.Request) {
	type DeleteRequest struct {
		Nim     string `json:"nim"`
		Kode_mk string `json:"kode_mk"`
	}

	var delRequest DeleteRequest

	err := json.NewDecoder(r.Body).Decode(&delRequest)
	if err != nil {
		http.Error(w, "JSON structure is not valid", 400)
		return
	}
	db, _ := getDB()
	defer db.Close()

	_, err = db.Exec("DELETE FROM perkuliahan WHERE kode_mk = ? AND nim = ?", delRequest.Kode_mk, delRequest.Nim)
	if err != nil {
		http.Error(w, "Delete failed", 500)
		return
	}
	replyHttpSuccessJson(w)
}

func handleNilaiMahasiswaPatch(w http.ResponseWriter, r *http.Request) {
	type PatchNilaiReq struct {
		Nim     string `json:"nim"`
		Kode_mk string `json:"kode_mk"`
		Nilai   int    `json:"nilai"`
	}

	var patchRequest PatchNilaiReq
	err := json.NewDecoder(r.Body).Decode(&patchRequest)
	if err != nil {
		http.Error(w, "JSON structure is not valid", 400)
		return
	}

	db, _ := getDB()
	defer db.Close()
	_, err = db.Exec("UPDATE perkuliahan SET nilai=? WHERE kode_mk=? AND nim=?",
		patchRequest.Nilai, patchRequest.Kode_mk, patchRequest.Nim)
	if err != nil {
		http.Error(w, "Update data failed", 501)
		return
	}
	replyHttpSuccessJson(w)
}

func handleNilaiMahasiswaPost(w http.ResponseWriter, r *http.Request) {
	type AddNilaiReq struct {
		Nim     string `json:"nim"`
		Kode_mk string `json:"kode_mk"`
		Nilai   int    `json:"nilai"`
	}

	var addNilaiReq AddNilaiReq
	err := json.NewDecoder(r.Body).Decode(&addNilaiReq)
	if err != nil {
		http.Error(w, "JSON structure is not valid", 400)
		return
	}

	db, _ := getDB()
	defer db.Close()
	_, err = db.Exec("INSERT INTO perkuliahan (nim, kode_mk, nilai) VALUES (?, ?, ?)",
		addNilaiReq.Nim, addNilaiReq.Kode_mk, addNilaiReq.Nilai)
	if err != nil {
		http.Error(w, "Insert data failed", 501)
		return
	}
	replyHttpSuccessJson(w)
}

func getNilaiMahasiswaJSON(nim string) string {
	rows, _ := queryDB(`SELECT mahasiswa.nim, nama, alamat, tanggal_lahir,
								   matakuliah.kode_mk, nama_mk, sks, nilai FROM mahasiswa
						JOIN perkuliahan ON perkuliahan.nim = mahasiswa.nim
						JOIN matakuliah ON matakuliah.kode_mk = perkuliahan.kode_mk
						WHERE mahasiswa.nim = ?`, nim)

	return jsonifyNilaiMahasiswaRows(rows)
}

func getNilaiMahasiswasJSON() string {
	rows, _ := queryDB(`SELECT mahasiswa.nim, nama, alamat, tanggal_lahir,
								   matakuliah.kode_mk, nama_mk, sks, nilai FROM mahasiswa
						JOIN perkuliahan ON perkuliahan.nim = mahasiswa.nim
						JOIN matakuliah ON matakuliah.kode_mk = perkuliahan.kode_mk`)

	return jsonifyNilaiMahasiswaRows(rows)
}

func jsonifyNilaiMahasiswaRows(rows *sql.Rows) string {
	resMap := map[string]map[string]interface{}{}

	type Grade struct {
		Kode_mk string `json:"kode_mk"`
		Nama_mk string `json:"nama_mk"`
		Sks     int    `json:"sks"`
		Nilai   int    `json:"nilai"`
	}

	for rows.Next() {
		var (
			nim           string
			nama          string
			alamat        string
			tanggal_lahir string
			kode_mk       string
			nama_mk       string
			sks           int
			nilai         int
		)
		if err := rows.Scan(&nim, &nama, &alamat, &tanggal_lahir, &kode_mk, &nama_mk, &sks, &nilai); err != nil {
			break
		}

		if _, ok := resMap[nim]; !ok {
			resMap[nim] = map[string]interface{}{
				"details": map[string]string{
					"nama":          nama,
					"alamat":        alamat,
					"tanggal_lahir": tanggal_lahir,
				},
				"grade": []Grade{}}
		}

		grade := resMap[nim]["grade"].([]Grade)
		resMap[nim]["grade"] = append(grade, Grade{
			kode_mk, nama_mk, sks, nilai})
	}

	jsonStr, _ := json.Marshal(resMap)

	return string(jsonStr)
}

func queryDB(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := getDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query(query, args...)

	return rows, err
}

func getDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", DBPATH)
	return db, err
}

func replyHttpSuccessJson(w http.ResponseWriter) {
	fmt.Fprintf(w, `{"status": "success"}`)
}
