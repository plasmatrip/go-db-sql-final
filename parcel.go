package main

import (
	"database/sql"
	"log"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// добавляем в таблицу parcel данные об отправлении
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	// возвращаем идентификатор добавленной строки
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// возвращаем информацию об отправлении по заданному трек-номеру
	p := Parcel{}

	row := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number", sql.Named("number", number))

	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	//выборка данных об отправлениях клиента по его ID
	//полученные данные помещаем в слайс
	var res []Parcel

	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client", sql.Named("client", client))
	if err != nil {
		log.Println(err)
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			log.Println(err)
			return res, err
		}
		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return res, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// обновляем статус отправления
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// изменям адрес отправления, при условии что статус registered
	_, err := s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number AND status = :status",
		sql.Named("address", address),
		sql.Named("status", ParcelStatusRegistered),
		sql.Named("number", number))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// удаляем информацию об отправлении, при условии что статус registered
	_, err := s.db.Exec("DELETE FROM parcel WHERE number = :number AND status = :status",
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
