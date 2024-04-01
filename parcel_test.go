package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// подключение БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	//создание объекта ParcelStore для работы с БД и тестового отправления
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавляем новую посылку в БД, ожидаем отсутствие ошибки и наличие идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// получаем только что добавленную посылку, ожидаем отсутствие ошибки
	// ожидаем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	p, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Address, p.Address)
	assert.Equal(t, parcel.Client, p.Client)
	assert.Equal(t, parcel.CreatedAt, p.CreatedAt)
	assert.Equal(t, parcel.Status, p.Status)

	// удвляем добавленную посылку, ожидаем отсутствие ошибки
	err = store.Delete(id)
	require.NoError(t, err)

	// ожидаем, что посылку больше нельзя получить из БД
	p, err = store.Get(id)
	require.Equal(t, sql.ErrNoRows, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// подключение БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	//создание объекта ParcelStore для работы с БД и тестового отправления
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавляем новую посылку в БД, ожидаем отсутствие ошибки и наличие идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// обновляем адрес, ожидаем отсутствие ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// получаем добавленную посылку и ожидаем, что адрес обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// подключение БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	//создание объекта ParcelStore для работы с БД и тестового отправления
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавляем новую посылку в БД, ожидаем отсутствие ошибки и наличие идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	// обновляем статус, ожидаем отсутствие ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// получаем добавленную посылку и ожидаем, что статус обновился
	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// подключение БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	//создание объекта ParcelStore для работы с БД и слайса тестовых отправлений
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// добавляем в цикле посылки в БД из слайса, ожидаем отсутствие ошибок и наличие идентификаторов
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// получаем список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// ожидаем отсутствие ошибки
	// ожидаем, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// ожидаем, что все посылки из storedParcels есть в parcelMap
	// ожидаем, что значения полей полученных посылок заполнены верно
	for _, parcel := range storedParcels {
		require.Equal(t, parcelMap[parcel.Number], parcel)
		assert.Equal(t, parcelMap[parcel.Number].Address, parcel.Address)
		assert.Equal(t, parcelMap[parcel.Number].Client, parcel.Client)
		assert.Equal(t, parcelMap[parcel.Number].CreatedAt, parcel.CreatedAt)
		assert.Equal(t, parcelMap[parcel.Number].Status, parcel.Status)
	}
}
